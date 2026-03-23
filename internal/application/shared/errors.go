package shared

import "errors"

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidPage        = errors.New("invalid page")
	ErrInvalidPageSize    = errors.New("invalid page size")
	ErrRoomNotFound       = errors.New("room not found")
	ErrScheduleExists     = errors.New("schedule already exists")
	ErrSlotNotFound       = errors.New("slot not found")
	ErrSlotBooked         = errors.New("slot already booked")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrDateTooFarAhead    = errors.New("requested date is too far ahead")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
)
