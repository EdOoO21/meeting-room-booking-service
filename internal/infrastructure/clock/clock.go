package clock

import (
	"time"
)

type Clock struct{}

func New() Clock {
	return Clock{}
}

func (Clock) NowUTC() time.Time {
	return time.Now().UTC()
}
