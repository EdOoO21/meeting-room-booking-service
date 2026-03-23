package http

import (
	appauth "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/auth"
	appbookings "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/bookings"
	approoms "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/rooms"
	appschedules "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/schedules"
	appslots "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/slots"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/ports"
)

type Services struct {
	Logger    ports.Logger
	Auth      *appauth.Service
	Rooms     *approoms.Service
	Schedules *appschedules.Service
	Slots     *appslots.Service
	Bookings  *appbookings.Service
}
