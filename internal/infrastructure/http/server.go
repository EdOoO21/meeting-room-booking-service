package http

import (
	"encoding/json"
	stdhttp "net/http"

	appauth "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/auth"
	appbookings "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/bookings"
	approoms "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/rooms"
	appschedules "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/schedules"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	appslots "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/slots"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const maxBookingsPageSize = 100

type Server struct {
	services Services
	generated.Unimplemented
}

func NewServer(services Services) *Server {
	return &Server{services: services}
}

// PostDummyLogin godoc
// @Summary Dummy login
// @Description Возвращает тестовый JWT для роли admin или user без регистрации.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SwaggerDummyLoginRequest true "Dummy login payload"
// @Success 200 {object} SwaggerTokenResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /dummyLogin [post]
func (s *Server) PostDummyLogin(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var body generated.PostDummyLoginJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: "invalid request body", Err: err})
		return
	}

	output, err := s.services.Auth.DummyLogin(r.Context(), appauth.DummyLoginInput{Role: domain.Role(body.Role)})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, generated.Token{Token: output.Token})
}

// PostRegister godoc
// @Summary Register user
// @Description Регистрирует нового пользователя по email и паролю.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SwaggerRegisterRequest true "Register payload"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 409 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /register [post]
func (s *Server) PostRegister(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var body generated.PostRegisterJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: "invalid request body", Err: err})
		return
	}

	output, err := s.services.Auth.Register(r.Context(), appauth.RegisterInput{
		Email:    string(body.Email),
		Password: body.Password,
		Role:     domain.Role(body.Role),
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusCreated, struct {
		User generated.User `json:"user"`
	}{User: mapUser(output.User)})
}

// PostLogin godoc
// @Summary Login
// @Description Выполняет обычный логин по email и паролю и возвращает JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SwaggerLoginRequest true "Login payload"
// @Success 200 {object} SwaggerTokenResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /login [post]
func (s *Server) PostLogin(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var body generated.PostLoginJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: "invalid request body", Err: err})
		return
	}

	output, err := s.services.Auth.Login(r.Context(), appauth.LoginInput{
		Email:    string(body.Email),
		Password: body.Password,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, generated.Token{Token: output.Token})
}

// GetRoomsList godoc
// @Summary List rooms
// @Description Возвращает список переговорок.
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Success 200 {object} RoomsListResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /rooms/list [get]
func (s *Server) GetRoomsList(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	output, err := s.services.Rooms.List(r.Context(), approoms.ListInput{Actor: actor})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	response := struct {
		Rooms []generated.Room `json:"rooms"`
	}{Rooms: make([]generated.Room, 0, len(output.Rooms))}

	for _, room := range output.Rooms {
		response.Rooms = append(response.Rooms, mapRoom(room))
	}

	writeJSON(w, stdhttp.StatusOK, response)
}

// PostRoomsCreate godoc
// @Summary Create room
// @Description Создает новую переговорку. Доступно только admin.
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SwaggerCreateRoomRequest true "Create room payload"
// @Success 201 {object} RoomResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 403 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /rooms/create [post]
func (s *Server) PostRoomsCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	var body generated.PostRoomsCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: "invalid request body", Err: err})
		return
	}

	output, err := s.services.Rooms.Create(r.Context(), approoms.CreateInput{
		Actor:       actor,
		Name:        body.Name,
		Description: stringOrEmpty(body.Description),
		Capacity:    body.Capacity,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusCreated, struct {
		Room generated.Room `json:"room"`
	}{Room: mapRoom(output.Room)})
}

// PostRoomsRoomIdScheduleCreate godoc
// @Summary Create room schedule
// @Description Создает расписание доступности для переговорки. Доступно только admin.
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room ID"
// @Param request body SwaggerCreateScheduleRequest true "Create schedule payload"
// @Success 201 {object} ScheduleResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 403 {object} SwaggerErrorResponse
// @Failure 404 {object} SwaggerErrorResponse
// @Failure 409 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /rooms/{roomId}/schedule/create [post]
//
//nolint:revive,staticcheck // Method name is fixed by generated oapi interface.
func (s *Server) PostRoomsRoomIdScheduleCreate(w stdhttp.ResponseWriter, r *stdhttp.Request, roomId generated.RoomIdPath) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	var body generated.PostRoomsRoomIdScheduleCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: "invalid request body", Err: err})
		return
	}

	days := make([]domain.DayOfWeek, 0, len(body.DaysOfWeek))
	for _, day := range body.DaysOfWeek {
		days = append(days, domain.DayOfWeek(day))
	}

	output, err := s.services.Schedules.Create(r.Context(), appschedules.CreateInput{
		Actor:      actor,
		RoomID:     roomId,
		DaysOfWeek: days,
		StartTime:  body.StartTime,
		EndTime:    body.EndTime,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusCreated, struct {
		Schedule generated.Schedule `json:"schedule"`
	}{Schedule: mapSchedule(output.Schedule)})
}

// GetRoomsRoomIdSlotsList godoc
// @Summary List available slots
// @Description Возвращает доступные для бронирования слоты по переговорке и дате.
// @Tags slots
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room ID"
// @Param date query string true "Date in YYYY-MM-DD format"
// @Success 200 {object} SlotsListResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 404 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /rooms/{roomId}/slots/list [get]
//
//nolint:revive,staticcheck // Method name is fixed by generated oapi interface.
func (s *Server) GetRoomsRoomIdSlotsList(w stdhttp.ResponseWriter, r *stdhttp.Request, roomId generated.RoomIdPath, params generated.GetRoomsRoomIdSlotsListParams) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	date := params.Date.Time.UTC()
	output, err := s.services.Slots.ListAvailable(r.Context(), appslots.ListAvailableInput{
		Actor:  actor,
		RoomID: roomId,
		Date:   date,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	response := struct {
		Slots []generated.Slot `json:"slots"`
	}{Slots: make([]generated.Slot, 0, len(output.Slots))}

	for _, slot := range output.Slots {
		response.Slots = append(response.Slots, mapSlot(slot))
	}

	writeJSON(w, stdhttp.StatusOK, response)
}

// PostBookingsCreate godoc
// @Summary Create booking
// @Description Создает бронь на слот. Опционально может запросить ссылку на конференцию. Доступно только user.
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SwaggerCreateBookingRequest true "Create booking payload"
// @Success 201 {object} BookingResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 403 {object} SwaggerErrorResponse
// @Failure 404 {object} SwaggerErrorResponse
// @Failure 409 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /bookings/create [post]
func (s *Server) PostBookingsCreate(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	var body generated.PostBookingsCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: "invalid request body", Err: err})
		return
	}

	createConferenceLink := body.CreateConferenceLink != nil && *body.CreateConferenceLink
	output, err := s.services.Bookings.Create(r.Context(), appbookings.CreateInput{
		Actor:                actor,
		SlotID:               body.SlotId,
		CreateConferenceLink: createConferenceLink,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusCreated, struct {
		Booking generated.Booking `json:"booking"`
	}{Booking: mapBooking(output.Booking)})
}

// PostBookingsBookingIdCancel godoc
// @Summary Cancel booking
// @Description Отменяет собственную бронь. Доступно только user.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param bookingId path string true "Booking ID"
// @Success 200 {object} BookingResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 403 {object} SwaggerErrorResponse
// @Failure 404 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /bookings/{bookingId}/cancel [post]
//
//nolint:revive,staticcheck // Method name is fixed by generated oapi interface.
func (s *Server) PostBookingsBookingIdCancel(w stdhttp.ResponseWriter, r *stdhttp.Request, bookingId generated.BookingIdPath) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	output, err := s.services.Bookings.Cancel(r.Context(), appbookings.CancelInput{
		Actor:     actor,
		BookingID: bookingId,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, struct {
		Booking generated.Booking `json:"booking"`
	}{Booking: mapBooking(output.Booking)})
}

// GetBookingsMy godoc
// @Summary List my bookings
// @Description Возвращает список броней текущего пользователя.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BookingsListResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 403 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /bookings/my [get]
func (s *Server) GetBookingsMy(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	output, err := s.services.Bookings.ListMine(r.Context(), appbookings.ListMineInput{Actor: actor})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	response := struct {
		Bookings []generated.Booking `json:"bookings"`
	}{Bookings: make([]generated.Booking, 0, len(output.Bookings))}

	for _, booking := range output.Bookings {
		response.Bookings = append(response.Bookings, mapBooking(booking))
	}

	writeJSON(w, stdhttp.StatusOK, response)
}

// GetBookingsList godoc
// @Summary List all bookings
// @Description Возвращает все брони с пагинацией. Доступно только admin.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size, max 100"
// @Success 200 {object} BookingsPageResponse
// @Failure 400 {object} SwaggerErrorResponse
// @Failure 401 {object} SwaggerErrorResponse
// @Failure 403 {object} SwaggerErrorResponse
// @Failure 500 {object} SwaggerErrorResponse
// @Router /bookings/list [get]
func (s *Server) GetBookingsList(w stdhttp.ResponseWriter, r *stdhttp.Request, params generated.GetBookingsListParams) {
	actor, ok := actorFromContext(r.Context())
	if !ok {
		writeAPIError(w, shared.ErrUnauthorized)
		return
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	pageSize := 20
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	if pageSize > maxBookingsPageSize {
		writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: "pageSize must not exceed 100"})
		return
	}

	output, err := s.services.Bookings.List(r.Context(), appbookings.ListInput{
		Actor:    actor,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	response := struct {
		Bookings   []generated.Booking  `json:"bookings"`
		Pagination generated.Pagination `json:"pagination"`
	}{
		Bookings: make([]generated.Booking, 0, len(output.Bookings)),
		Pagination: generated.Pagination{
			Page:     output.Pagination.Page,
			PageSize: output.Pagination.PageSize,
			Total:    output.Pagination.Total,
		},
	}

	for _, booking := range output.Bookings {
		response.Bookings = append(response.Bookings, mapBooking(booking))
	}

	writeJSON(w, stdhttp.StatusOK, response)
}

func mapUser(user domain.User) generated.User {
	response := generated.User{
		Id:    user.ID,
		Email: openapi_types.Email(user.Email),
		Role:  generated.UserRole(user.Role),
	}
	createdAt := user.CreatedAt.UTC()
	response.CreatedAt = &createdAt
	return response
}

func mapRoom(room domain.Room) generated.Room {
	response := generated.Room{
		Id:   room.ID,
		Name: room.Name,
	}

	if room.Description != "" {
		description := room.Description
		response.Description = &description
	}
	if room.Capacity != nil {
		capacity := *room.Capacity
		response.Capacity = &capacity
	}
	createdAt := room.CreatedAt.UTC()
	response.CreatedAt = &createdAt

	return response
}

func mapSchedule(schedule domain.Schedule) generated.Schedule {
	response := generated.Schedule{
		RoomId:     schedule.RoomID,
		DaysOfWeek: make([]int, 0, len(schedule.DaysOfWeek)),
		StartTime:  schedule.StartTime.String(),
		EndTime:    schedule.EndTime.String(),
	}
	id := schedule.ID
	response.Id = &id
	for _, day := range schedule.DaysOfWeek {
		response.DaysOfWeek = append(response.DaysOfWeek, int(day))
	}
	return response
}

func mapSlot(slot domain.Slot) generated.Slot {
	return generated.Slot{
		Id:     slot.ID,
		RoomId: slot.RoomID,
		Start:  slot.Start.UTC(),
		End:    slot.End.UTC(),
	}
}

func mapBooking(booking domain.Booking) generated.Booking {
	response := generated.Booking{
		Id:     booking.ID,
		SlotId: booking.SlotID,
		UserId: booking.UserID,
		Status: generated.BookingStatus(booking.Status),
	}
	if booking.ConferenceLink != nil {
		link := *booking.ConferenceLink
		response.ConferenceLink = &link
	}
	createdAt := booking.CreatedAt.UTC()
	response.CreatedAt = &createdAt
	return response
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
