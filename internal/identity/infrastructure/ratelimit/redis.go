package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisTTLMultiplier — TTL ключей = window * N, чтобы покрыть текущее и предыдущее окно.
const redisTTLMultiplier = 2

// RedisLimiter реализует Sliding Window Counter.
// Используются два соседних окна с взвешенным подсчётом,
// что устраняет проблему burst на границе fixed window.
type RedisLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
	prefix string
}

func NewRedisLimiter(client *redis.Client, limit int, window time.Duration, prefix string) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		limit:  limit,
		window: window,
		prefix: prefix,
	}
}

// slidingWindowCounter — Lua-скрипт, выполняемый атомарно в Redis.
// Считает взвешенную сумму запросов из текущего и предыдущего окна:
//   estimate = prev_count * (1 - elapsed/window) + curr_count
var slidingWindowCounter = redis.NewScript(`
local prev_key  = KEYS[1]
local curr_key  = KEYS[2]
local limit     = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local now_ms    = tonumber(ARGV[3])
local ttl_mult  = tonumber(ARGV[4])

local curr_start = math.floor(now_ms / window_ms) * window_ms
local elapsed    = now_ms - curr_start
local weight     = 1 - (elapsed / window_ms)

local prev_count = tonumber(redis.call('GET', prev_key) or '0') or 0
local curr_count = tonumber(redis.call('GET', curr_key) or '0') or 0

local estimate = prev_count * weight + curr_count
if estimate >= limit then
    return 0
end

redis.call('INCR', curr_key)
redis.call('PEXPIRE', curr_key, window_ms * ttl_mult)
redis.call('PEXPIRE', prev_key, window_ms * ttl_mult)
return 1
`)

func (r *RedisLimiter) Allow(key string) bool {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	windowMs := r.window.Milliseconds()
	currStart := (now / windowMs) * windowMs
	prevStart := currStart - windowMs

	prevKey := fmt.Sprintf("ratelimit:%s:%s:%d", r.prefix, key, prevStart)
	currKey := fmt.Sprintf("ratelimit:%s:%s:%d", r.prefix, key, currStart)

	result, err := slidingWindowCounter.Run(ctx, r.client,
		[]string{prevKey, currKey},
		r.limit, windowMs, now, redisTTLMultiplier,
	).Int()
	if err != nil {
		return true
	}
	return result == 1
}

func (r *RedisLimiter) Reset(key string) {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	windowMs := r.window.Milliseconds()
	currStart := (now / windowMs) * windowMs
	prevStart := currStart - windowMs

	prevKey := fmt.Sprintf("ratelimit:%s:%s:%d", r.prefix, key, prevStart)
	currKey := fmt.Sprintf("ratelimit:%s:%s:%d", r.prefix, key, currStart)
	r.client.Del(ctx, prevKey, currKey)
}
