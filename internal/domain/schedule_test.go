package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSchedule_NormalizesUniqueSortedDays(t *testing.T) {
	t.Parallel()

	schedule, err := NewSchedule(
		uuid.New(),
		uuid.New(),
		[]DayOfWeek{Friday, Monday, Friday, Wednesday},
		TimeOfDay{Hour: 9, Minute: 0},
		TimeOfDay{Hour: 11, Minute: 0},
	)
	if err != nil {
		t.Fatalf("NewSchedule() error = %v", err)
	}

	wantDays := []DayOfWeek{Monday, Wednesday, Friday}
	if !reflect.DeepEqual(schedule.DaysOfWeek, wantDays) {
		t.Fatalf("schedule.DaysOfWeek = %v, want %v", schedule.DaysOfWeek, wantDays)
	}
}

func TestNewSchedule_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	validStart := TimeOfDay{Hour: 9, Minute: 0}
	validEnd := TimeOfDay{Hour: 10, Minute: 0}

	tests := []struct {
		name      string
		id        uuid.UUID
		roomID    uuid.UUID
		days      []DayOfWeek
		startTime TimeOfDay
		endTime   TimeOfDay
		want      error
	}{
		{name: "invalid schedule id", id: uuid.Nil, roomID: uuid.New(), days: []DayOfWeek{Monday}, startTime: validStart, endTime: validEnd, want: ErrInvalidID},
		{name: "invalid room id", id: uuid.New(), roomID: uuid.Nil, days: []DayOfWeek{Monday}, startTime: validStart, endTime: validEnd, want: ErrInvalidID},
		{name: "days required", id: uuid.New(), roomID: uuid.New(), days: nil, startTime: validStart, endTime: validEnd, want: ErrScheduleDaysRequired},
		{name: "invalid weekday", id: uuid.New(), roomID: uuid.New(), days: []DayOfWeek{DayOfWeek(9)}, startTime: validStart, endTime: validEnd, want: ErrInvalidDayOfWeek},
		{name: "invalid start time", id: uuid.New(), roomID: uuid.New(), days: []DayOfWeek{Monday}, startTime: TimeOfDay{Hour: 25}, endTime: validEnd, want: ErrInvalidTimeOfDay},
		{name: "invalid end time", id: uuid.New(), roomID: uuid.New(), days: []DayOfWeek{Monday}, startTime: validStart, endTime: TimeOfDay{Minute: 70}, want: ErrInvalidTimeOfDay},
		{name: "start after end", id: uuid.New(), roomID: uuid.New(), days: []DayOfWeek{Monday}, startTime: TimeOfDay{Hour: 10}, endTime: TimeOfDay{Hour: 9}, want: ErrScheduleTimeRangeOrder},
		{name: "too short", id: uuid.New(), roomID: uuid.New(), days: []DayOfWeek{Monday}, startTime: TimeOfDay{Hour: 9}, endTime: TimeOfDay{Hour: 9, Minute: 15}, want: ErrScheduleTimeRangeTooShort},
		{name: "not aligned", id: uuid.New(), roomID: uuid.New(), days: []DayOfWeek{Monday}, startTime: TimeOfDay{Hour: 9}, endTime: TimeOfDay{Hour: 9, Minute: 45}, want: ErrScheduleTimeRangeNotAligned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewSchedule(tt.id, tt.roomID, tt.days, tt.startTime, tt.endTime)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewSchedule() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseTimeOfDay(t *testing.T) {
	t.Parallel()

	parsed, err := ParseTimeOfDay("09:30")
	if err != nil {
		t.Fatalf("ParseTimeOfDay() error = %v", err)
	}

	if parsed != (TimeOfDay{Hour: 9, Minute: 30}) {
		t.Fatalf("parsed = %+v, want 09:30", parsed)
	}

	if parsed.String() != "09:30" {
		t.Fatalf("parsed.String() = %q, want %q", parsed.String(), "09:30")
	}

	if _, parseErr := ParseTimeOfDay("24:00"); !errors.Is(parseErr, ErrInvalidTimeOfDay) {
		t.Fatalf("ParseTimeOfDay() error = %v, want %v", parseErr, ErrInvalidTimeOfDay)
	}
}

func TestDayOfWeekMapping(t *testing.T) {
	t.Parallel()

	if got := DayOfWeekFromWeekday(time.Monday); got != Monday {
		t.Fatalf("DayOfWeekFromWeekday(Monday) = %v, want %v", got, Monday)
	}

	if got := Friday.ToTimeWeekday(); got != time.Friday {
		t.Fatalf("Friday.ToTimeWeekday() = %v, want %v", got, time.Friday)
	}
}
