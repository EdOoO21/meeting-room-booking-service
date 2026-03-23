package domain

import "time"

func normalizeUTC(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}

	return value.UTC()
}

func RequireUTC(value time.Time) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, ErrInvalidTimestamp
	}

	if value.Location() != time.UTC {
		return time.Time{}, ErrNonUTCTimestamp
	}

	return value, nil
}
