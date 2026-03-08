package domain

type RateLimiter interface {
	Allow(key string) bool
	Reset(key string)
}
