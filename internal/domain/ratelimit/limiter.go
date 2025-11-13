package ratelimit

import (
	"context"
	"time"
)

// Rule описывает правило для одного типа лимита (например, login или ip).
// Пример: Rule{Limit: 10, Window: time.Minute}.
type Rule struct {
	Limit  int
	Window time.Duration
}

// Limiter — реализация по алгоритму скользящего окна.
// Для каждого типа лимита (limitType) отдельный экземпляр Limiter.
type Limiter struct {
	storage   LimiterRepo
	limitType string // тип лимита (login, password или ip)
	rule      Rule
}

// NewLimiter возвращает новую in-memory реализацию.
func NewLimiter(st LimiterRepo, limitType string, r Rule) *Limiter {
	return &Limiter{
		storage:   st,
		limitType: limitType,
		rule:      r,
	}
}

// Allow проверяет, разрешена ли по лимитеру попытка для данного ключа.
func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	key = l.limitType + ":" + key
	return l.storage.Allow(ctx, key, l.rule.Limit, l.rule.Window)
}

// Reset сбрасывает счетчик попыток для данного лимитера для данного ключа.
func (l *Limiter) Reset(ctx context.Context, key string) error {
	key = l.limitType + ":" + key
	return l.storage.Reset(ctx, key)
}
