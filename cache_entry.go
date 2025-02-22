package bmemcache

import "time"

type cacheEntry[T any] struct {
	Data T
	Exp  time.Time
}

func (ce *cacheEntry[T]) isExpired() bool {
	return !ce.Exp.IsZero() && time.Now().After(ce.Exp)
}

func (ce *cacheEntry[T]) flush() {
	ce.Data = generateEmptyData[T]()
}
