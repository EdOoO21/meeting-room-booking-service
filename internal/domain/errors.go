package domain

import "errors"

var (
	ErrInvalidID                   = errors.New("invalid id")
	ErrInvalidEmail                = errors.New("invalid email")
	ErrInvalidRole                 = errors.New("invalid role")
	ErrInvalidRoomName             = errors.New("invalid room name")
	ErrInvalidRoomCapacity         = errors.New("invalid room capacity")
	ErrInvalidSchedule             = errors.New("invalid schedule")
	ErrScheduleDaysRequired        = errors.New("schedule days of week are required")
	ErrInvalidDayOfWeek            = errors.New("invalid day of week")
	ErrInvalidTimeOfDay            = errors.New("invalid time of day")
	ErrScheduleTimeRangeOrder      = errors.New("schedule start time must be before end time")
	ErrScheduleTimeRangeTooShort   = errors.New("schedule time range must fit at least one slot")
	ErrScheduleTimeRangeNotAligned = errors.New("schedule time range must be divisible by slot duration")
	ErrInvalidTimestamp            = errors.New("invalid timestamp")
	ErrNonUTCTimestamp             = errors.New("timestamp must be in utc")
	ErrInvalidSlot                 = errors.New("invalid slot")
	ErrSlotTimeRangeOrder          = errors.New("slot start time must be before end time")
	ErrSlotDurationMustMatch       = errors.New("slot duration must equal fixed slot duration")
	ErrInvalidBooking              = errors.New("invalid booking")
	ErrInvalidBookingStatus        = errors.New("invalid booking status")
	ErrPastSlotBooking             = errors.New("cannot book a slot in the past")
	ErrForbiddenBookingByRole      = errors.New("role is not allowed to create bookings")
)
