package ratelimit

import (
	"context"
	"time"
)

// LimiterRepo — абстракция ограничения.
// Важно: реализация должна быть атомарной в разрезе одного key.
type LimiterRepo interface {
	// Allow применяет скользящее окно к ОДНОМУ ключу.
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	// Reset очищает окно для ключа (для метода "сброс bucket").
	Reset(ctx context.Context, key string) error
}
