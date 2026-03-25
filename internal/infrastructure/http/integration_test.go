package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appauth "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/auth"
	appbookings "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/bookings"
	approoms "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/rooms"
	appschedules "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/schedules"
	appslots "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/slots"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	appclock "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/clock"
	appconference "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/conference"
	httptransport "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http/generated"
	appid "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/id"
	appjwt "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/jwt"
	logs "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/logger"
	apppassword "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/password"
	apppostgres "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestIntegration_CreateRoomScheduleAndBooking(t *testing.T) {
	app := newIntegrationApp(t)

	adminToken := loginByRole(t, app.router, "admin")
	userToken := loginByRole(t, app.router, "user")

	room := createRoom(t, app.router, adminToken)
	targetDate := time.Now().UTC().AddDate(0, 0, 1)

	createSchedule(t, app.router, adminToken, room.Id, targetDate)
	slots := listSlots(t, app.router, userToken, room.Id, targetDate)
	if len(slots) == 0 {
		t.Fatal("expected at least one available slot")
	}

	booking := createBooking(t, app.router, userToken, slots[0].Id, true)
	if booking.Status != generated.Active {
		t.Fatalf("expected active booking, got %q", booking.Status)
	}
	if booking.ConferenceLink == nil || *booking.ConferenceLink == "" {
		t.Fatal("expected conference link to be created")
	}
}

func TestIntegration_RegisterLoginAndListFlows(t *testing.T) {
	app := newIntegrationApp(t)

	unauthorized := performJSONRequest(t, app.router, http.MethodGet, "/rooms/list", nil, "")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("rooms list without token status = %d, want %d; body = %s", unauthorized.Code, http.StatusUnauthorized, unauthorized.Body.String())
	}

	registeredToken := registerUser(t, app.router, "tester@example.com", "strong-password-123", "user")
	rooms := listRooms(t, app.router, registeredToken)
	if len(rooms) != 0 {
		t.Fatalf("expected no rooms for fresh database, got %d", len(rooms))
	}

	loginToken := loginUser(t, app.router, "tester@example.com", "strong-password-123")
	rooms = listRooms(t, app.router, loginToken)
	if len(rooms) != 0 {
		t.Fatalf("expected no rooms after login on fresh database, got %d", len(rooms))
	}

	adminToken := loginByRole(t, app.router, "admin")
	room := createRoom(t, app.router, adminToken)
	targetDate := time.Now().UTC().AddDate(0, 0, 1)
	createSchedule(t, app.router, adminToken, room.Id, targetDate)

	slots := listSlots(t, app.router, loginToken, room.Id, targetDate)
	if len(slots) == 0 {
		t.Fatal("expected at least one available slot")
	}

	created := createBooking(t, app.router, loginToken, slots[0].Id, false)
	myBookings := listMyBookings(t, app.router, loginToken)
	if len(myBookings) != 1 {
		t.Fatalf("expected 1 booking in my bookings, got %d", len(myBookings))
	}
	if myBookings[0].Id != created.Id {
		t.Fatalf("my bookings returned %s, want %s", myBookings[0].Id, created.Id)
	}

	allBookings, pagination := listAllBookings(t, app.router, adminToken)
	if len(allBookings) != 1 {
		t.Fatalf("expected 1 booking in admin list, got %d", len(allBookings))
	}
	if allBookings[0].Id != created.Id {
		t.Fatalf("admin bookings returned %s, want %s", allBookings[0].Id, created.Id)
	}
	if pagination.Total != 1 {
		t.Fatalf("admin bookings total = %d, want 1", pagination.Total)
	}
}

func TestIntegration_CancelBookingByUser(t *testing.T) {
	app := newIntegrationApp(t)

	adminToken := loginByRole(t, app.router, "admin")
	userToken := loginByRole(t, app.router, "user")

	room := createRoom(t, app.router, adminToken)
	targetDate := time.Now().UTC().AddDate(0, 0, 1)

	createSchedule(t, app.router, adminToken, room.Id, targetDate)
	slots := listSlots(t, app.router, userToken, room.Id, targetDate)
	if len(slots) == 0 {
		t.Fatal("expected at least one available slot")
	}

	created := createBooking(t, app.router, userToken, slots[0].Id, false)
	cancelled := cancelBooking(t, app.router, userToken, created.Id)
	if cancelled.Status != generated.Cancelled {
		t.Fatalf("expected cancelled booking, got %q", cancelled.Status)
	}

	cancelledAgain := cancelBooking(t, app.router, userToken, created.Id)
	if cancelledAgain.Status != generated.Cancelled {
		t.Fatalf("expected cancelled booking after second cancel, got %q", cancelledAgain.Status)
	}
}

type integrationApp struct {
	router http.Handler
}

func newIntegrationApp(t *testing.T) integrationApp {
	t.Helper()

	dsn := testDSN(t)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("http integration tests skipped: create pg pool: %v", err)
	}
	defer pool.Close()

	if pingErr := pool.Ping(ctx); pingErr != nil {
		t.Skipf("http integration tests skipped: ping pg: %v", pingErr)
	}

	resetTestDatabase(ctx, t, pool)

	db, err := apppostgres.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect postgres wrapper: %v", err)
	}
	t.Cleanup(db.Close)

	clock := appclock.New()
	ids := appid.New()
	passwords := apppassword.New()
	tokens := appjwt.New("integration-test-secret", time.Hour)
	conferenceLinks := appconference.NewMock()
	txManager := apppostgres.NewTxManager(db)

	userRepo := apppostgres.NewUserRepository(db)
	roomRepo := apppostgres.NewRoomRepository(db)
	scheduleRepo := apppostgres.NewScheduleRepository(db)
	slotRepo := apppostgres.NewSlotRepository(db)
	bookingRepo := apppostgres.NewBookingRepository(db)

	services := httptransport.Services{
		Logger:    logs.NewLogger(),
		JWT:       tokens,
		Auth:      appauth.NewService(userRepo, ids, clock, passwords, tokens),
		Rooms:     approoms.NewService(roomRepo, ids, clock),
		Schedules: appschedules.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock),
		Slots:     appslots.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock),
		Bookings:  appbookings.NewService(bookingRepo, slotRepo, txManager, ids, clock, conferenceLinks),
	}

	return integrationApp{router: httptransport.NewRouter(services)}
}

func testDSN(t *testing.T) string {
	t.Helper()

	if dsn := strings.TrimSpace(os.Getenv("APP_POSTGRES_TEST_DSN")); dsn != "" {
		return dsn
	}

	envPath := filepath.Join("..", "..", "..", ".env")
	data, err := os.ReadFile(envPath)
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if ok && strings.TrimSpace(key) == "APP_POSTGRES_TEST_DSN" {
				return strings.TrimSpace(value)
			}
		}
	}

	t.Skip("APP_POSTGRES_TEST_DSN is not set")
	return ""
}

func resetTestDatabase(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatalf("reset schema: %v", err)
	}

	migrationPath := filepath.Join("..", "..", "..", "migrations", "000001_init.up.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	for _, statement := range splitSQLStatements(string(migrationSQL)) {
		if _, execErr := pool.Exec(ctx, statement); execErr != nil {
			t.Fatalf("apply migration statement %q: %v", statement, execErr)
		}
	}
}

func splitSQLStatements(migration string) []string {
	lines := strings.Split(migration, "\n")
	statements := make([]string, 0)
	var current strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)

		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

func registerUser(t *testing.T, handler http.Handler, email, password, role string) string {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodPost, "/register", map[string]any{
		"email":    email,
		"password": password,
		"role":     role,
	}, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		User generated.User `json:"user"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)
	if string(payload.User.Email) != email {
		t.Fatalf("registered email = %q, want %q", payload.User.Email, email)
	}
	if payload.User.Role != generated.UserRole(role) {
		t.Fatalf("registered role = %q, want %q", payload.User.Role, role)
	}
	if payload.User.CreatedAt == nil {
		t.Fatal("expected createdAt in register response")
	}

	return loginUser(t, handler, email, password)
}

func loginUser(t *testing.T, handler http.Handler, email, password string) string {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodPost, "/login", map[string]any{
		"email":    email,
		"password": password,
	}, "")
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Token string `json:"token"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)
	if payload.Token == "" {
		t.Fatal("expected jwt token from login")
	}

	return payload.Token
}

func listMyBookings(t *testing.T, handler http.Handler, token string) []generated.Booking {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodGet, "/bookings/my", nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("my bookings status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Bookings []generated.Booking `json:"bookings"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)

	return payload.Bookings
}

func listAllBookings(t *testing.T, handler http.Handler, token string) ([]generated.Booking, generated.Pagination) {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodGet, "/bookings/list", nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("bookings list status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Bookings   []generated.Booking  `json:"bookings"`
		Pagination generated.Pagination `json:"pagination"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)

	return payload.Bookings, payload.Pagination
}

func listRooms(t *testing.T, handler http.Handler, token string) []generated.Room {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodGet, "/rooms/list", nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("rooms list status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Rooms []generated.Room `json:"rooms"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)

	return payload.Rooms
}

func loginByRole(t *testing.T, handler http.Handler, role string) string {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodPost, "/dummyLogin", map[string]any{"role": role}, "")
	if response.Code != http.StatusOK {
		t.Fatalf("dummy login status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Token string `json:"token"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)
	if payload.Token == "" {
		t.Fatal("expected jwt token")
	}

	return payload.Token
}

func createRoom(t *testing.T, handler http.Handler, token string) generated.Room {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodPost, "/rooms/create", map[string]any{
		"name":        "Focus Room",
		"description": "quiet room",
		"capacity":    6,
	}, token)
	if response.Code != http.StatusCreated {
		t.Fatalf("create room status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Room generated.Room `json:"room"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)

	return payload.Room
}

func createSchedule(t *testing.T, handler http.Handler, token string, roomID openapi_types.UUID, targetDate time.Time) {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", map[string]any{
		"roomId":     roomID.String(),
		"daysOfWeek": []int{int(domain.DayOfWeekFromWeekday(targetDate.Weekday()))},
		"startTime":  "09:00",
		"endTime":    "10:00",
	}, token)
	if response.Code != http.StatusCreated {
		t.Fatalf("create schedule status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Schedule generated.Schedule `json:"schedule"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)
}

func listSlots(t *testing.T, handler http.Handler, token string, roomID openapi_types.UUID, targetDate time.Time) []generated.Slot {
	t.Helper()

	path := "/rooms/" + roomID.String() + "/slots/list?date=" + targetDate.UTC().Format("2006-01-02")
	response := performJSONRequest(t, handler, http.MethodGet, path, nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("list slots status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Slots []generated.Slot `json:"slots"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)

	return payload.Slots
}

func createBooking(t *testing.T, handler http.Handler, token string, slotID openapi_types.UUID, createConferenceLink bool) generated.Booking {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodPost, "/bookings/create", map[string]any{
		"slotId":               slotID.String(),
		"createConferenceLink": createConferenceLink,
	}, token)
	if response.Code != http.StatusCreated {
		t.Fatalf("create booking status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Booking generated.Booking `json:"booking"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)

	return payload.Booking
}

func cancelBooking(t *testing.T, handler http.Handler, token string, bookingID openapi_types.UUID) generated.Booking {
	t.Helper()

	response := performJSONRequest(t, handler, http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("cancel booking status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Booking generated.Booking `json:"booking"`
	}
	decodeJSON(t, response.Body.Bytes(), &payload)

	return payload.Booking
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decodeJSON(t *testing.T, data []byte, target any) {
	t.Helper()

	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode response body: %v; body = %s", err, string(data))
	}
}
