package middlewares

import (
	"context"
	"time"

	"walki/internal/handlers/mux"
)

// Оборачивает хэндлер таймаутом (по умолчанию 5с)
func Timeout(d time.Duration) mux.Middleware {
	return func(next mux.HandlerFunc) mux.HandlerFunc {
		return func(u *mux.UpdateCtx) error {
			ctx, cancel := context.WithTimeout(u.Ctx, d)
			defer cancel()
			u.Ctx = ctx
			return next(u)
		}
	}
}
