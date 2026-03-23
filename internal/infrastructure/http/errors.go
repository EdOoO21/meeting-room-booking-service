package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http/generated"
)

type apiError struct {
	Status  int
	Code    generated.ErrorResponseErrorCode
	Message string
	Err     error
}

func (e apiError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func writeAPIError(w stdhttp.ResponseWriter, err error) {
	mapped := mapAPIError(err)

	response := generated.ErrorResponse{}
	response.Error.Code = mapped.Code
	response.Error.Message = mapped.Message

	writeJSON(w, mapped.Status, response)
}

func mapAPIError(err error) apiError {
	switch {
	case errors.Is(err, shared.ErrUnauthorized):
		return apiError{Status: stdhttp.StatusUnauthorized, Code: generated.UNAUTHORIZED, Message: "unauthorized", Err: err}
	case errors.Is(err, shared.ErrInvalidCredentials):
		return apiError{Status: stdhttp.StatusUnauthorized, Code: generated.UNAUTHORIZED, Message: "invalid credentials", Err: err}
	case errors.Is(err, shared.ErrForbidden), errors.Is(err, domain.ErrForbiddenBookingByRole):
		return apiError{Status: stdhttp.StatusForbidden, Code: generated.FORBIDDEN, Message: err.Error(), Err: err}
	case errors.Is(err, shared.ErrRoomNotFound):
		return apiError{Status: stdhttp.StatusNotFound, Code: generated.ROOMNOTFOUND, Message: "room not found", Err: err}
	case errors.Is(err, shared.ErrSlotNotFound):
		return apiError{Status: stdhttp.StatusNotFound, Code: generated.SLOTNOTFOUND, Message: "slot not found", Err: err}
	case errors.Is(err, shared.ErrBookingNotFound):
		return apiError{Status: stdhttp.StatusNotFound, Code: generated.BOOKINGNOTFOUND, Message: "booking not found", Err: err}
	case errors.Is(err, shared.ErrScheduleExists):
		return apiError{Status: stdhttp.StatusConflict, Code: generated.SCHEDULEEXISTS, Message: "schedule for this room already exists and cannot be changed", Err: err}
	case errors.Is(err, shared.ErrSlotBooked):
		return apiError{Status: stdhttp.StatusConflict, Code: generated.SLOTALREADYBOOKED, Message: "slot is already booked", Err: err}
	case errors.Is(err, shared.ErrEmailAlreadyExists), errors.Is(err, shared.ErrInvalidPassword), errors.Is(err, shared.ErrInvalidPage), errors.Is(err, shared.ErrInvalidPageSize), errors.Is(err, shared.ErrDateTooFarAhead):
		return apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: err.Error(), Err: err}
	case errors.Is(err, domain.ErrInvalidRole), errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidRoomName), errors.Is(err, domain.ErrInvalidRoomCapacity), errors.Is(err, domain.ErrInvalidDayOfWeek), errors.Is(err, domain.ErrInvalidTimeOfDay), errors.Is(err, domain.ErrScheduleDaysRequired), errors.Is(err, domain.ErrScheduleTimeRangeOrder), errors.Is(err, domain.ErrScheduleTimeRangeTooShort), errors.Is(err, domain.ErrScheduleTimeRangeNotAligned), errors.Is(err, domain.ErrInvalidTimestamp), errors.Is(err, domain.ErrNonUTCTimestamp), errors.Is(err, domain.ErrPastSlotBooking):
		return apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: err.Error(), Err: err}
	default:
		return apiError{Status: stdhttp.StatusInternalServerError, Code: generated.INTERNALERROR, Message: "internal server error", Err: err}
	}
}

func writeGeneratedParamError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	writeAPIError(w, apiError{Status: stdhttp.StatusBadRequest, Code: generated.INVALIDREQUEST, Message: err.Error(), Err: err})
}

func writeJSON(w stdhttp.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		stdhttp.Error(w, stdhttp.StatusText(stdhttp.StatusInternalServerError), stdhttp.StatusInternalServerError)
	}
}
