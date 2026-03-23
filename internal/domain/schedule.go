package domain

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

const SlotDuration = 30 * time.Minute

type DayOfWeek int

const (
	Monday DayOfWeek = 1 + iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

type TimeOfDay struct {
	Hour   int
	Minute int
}

type Schedule struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	DaysOfWeek []DayOfWeek
	StartTime  TimeOfDay
	EndTime    TimeOfDay
}

func NewSchedule(id, roomID uuid.UUID, daysOfWeek []DayOfWeek, startTime, endTime TimeOfDay) (Schedule, error) {
	if id == uuid.Nil || roomID == uuid.Nil {
		return Schedule{}, ErrInvalidID
	}

	validatedDays, err := normalizeDaysOfWeek(daysOfWeek)
	if err != nil {
		return Schedule{}, err
	}

	if !startTime.IsValid() || !endTime.IsValid() {
		return Schedule{}, ErrInvalidTimeOfDay
	}

	rangeDuration := endTime.ToDuration() - startTime.ToDuration()

	if !startTime.Before(endTime) {
		return Schedule{}, ErrScheduleTimeRangeOrder
	}

	if rangeDuration < SlotDuration {
		return Schedule{}, ErrScheduleTimeRangeTooShort
	}

	if rangeDuration%SlotDuration != 0 {
		return Schedule{}, ErrScheduleTimeRangeNotAligned
	}

	return Schedule{
		ID:         id,
		RoomID:     roomID,
		DaysOfWeek: validatedDays,
		StartTime:  startTime,
		EndTime:    endTime,
	}, nil
}

func ParseTimeOfDay(value string) (TimeOfDay, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return TimeOfDay{}, fmt.Errorf("%w: %v", ErrInvalidTimeOfDay, err)
	}

	result := TimeOfDay{
		Hour:   parsed.Hour(),
		Minute: parsed.Minute(),
	}

	if !result.IsValid() {
		return TimeOfDay{}, ErrInvalidTimeOfDay
	}

	return result, nil
}

func (d DayOfWeek) IsValid() bool {
	return d >= Monday && d <= Sunday
}

func (d DayOfWeek) ToTimeWeekday() time.Weekday {
	switch d {
	case Monday:
		return time.Monday
	case Tuesday:
		return time.Tuesday
	case Wednesday:
		return time.Wednesday
	case Thursday:
		return time.Thursday
	case Friday:
		return time.Friday
	case Saturday:
		return time.Saturday
	case Sunday:
		return time.Sunday
	default:
		return time.Sunday
	}
}

func DayOfWeekFromWeekday(weekday time.Weekday) DayOfWeek {
	switch weekday {
	case time.Monday:
		return Monday
	case time.Tuesday:
		return Tuesday
	case time.Wednesday:
		return Wednesday
	case time.Thursday:
		return Thursday
	case time.Friday:
		return Friday
	case time.Saturday:
		return Saturday
	default:
		return Sunday
	}
}

func (t TimeOfDay) IsValid() bool {
	return t.Hour >= 0 && t.Hour <= 23 && t.Minute >= 0 && t.Minute <= 59
}

func (t TimeOfDay) Before(other TimeOfDay) bool {
	return t.toMinuteOfDay() < other.toMinuteOfDay()
}

func (t TimeOfDay) ToDuration() time.Duration {
	return time.Duration(t.toMinuteOfDay()) * time.Minute
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

func (t TimeOfDay) toMinuteOfDay() int {
	return t.Hour*60 + t.Minute
}

func normalizeDaysOfWeek(days []DayOfWeek) ([]DayOfWeek, error) {
	if len(days) == 0 {
		return nil, ErrScheduleDaysRequired
	}

	seen := make(map[DayOfWeek]struct{}, len(days))
	validated := make([]DayOfWeek, 0, len(days))

	for _, day := range days {
		if !day.IsValid() {
			return nil, ErrInvalidDayOfWeek
		}

		if _, exists := seen[day]; exists {
			continue
		}

		seen[day] = struct{}{}
		validated = append(validated, day)
	}

	slices.Sort(validated)

	return validated, nil
}
