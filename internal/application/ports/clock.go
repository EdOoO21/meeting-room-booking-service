package ports

import "time"

type Clock interface {
	NowUTC() time.Time
}
