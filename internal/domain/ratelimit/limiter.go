package ratelimit

import (
	"context"
	"time"
)

// Rule описывает один лимит для одного ключа (например, login или ip).
// Пример: Rule{Limit: 10, Window: time.Minute}.
type Rule struct {
	Limit  int
	Window time.Duration
}

// Limiter — реализация по алгоритму скользящего окна.
type Limiter struct {
	db   LimiterStorage
	key  string
	rule Rule
}

// NewInMemoryLimiter возвращает новую in-memory реализацию.
func NewInMemoryLimiter(st LimiterStorage, key string, r Rule) *Limiter {
	return &Limiter{
		db:   st,
		key:  key,
		rule: r,
	}
}

// Allow проверяет, разрешена ли попытка для данного лимитера.
func (l *Limiter) Allow(ctx context.Context) (bool, error) {
	return l.db.Allow(ctx, l.key, l.rule.Limit, l.rule.Window)
}

// Reset сбрасывает счетчик попыток для данного лимитера.
func (l *Limiter) Reset(ctx context.Context) error {
	return l.db.Reset(ctx, l.key)
}
