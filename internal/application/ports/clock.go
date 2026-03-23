package ports

import "time"

// Clock предоставляет текущее время для usecases
type Clock interface {
	// NowUTC возвращает текущее время в UTC
	NowUTC() time.Time
}
