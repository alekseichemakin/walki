package middlewares

import (
	"context"
	"walki/internal/handlers/mux"
	"walki/internal/models"
	"walki/internal/service"
)

type userKeyType struct{}

var userKey userKeyType

func WithUser(users *service.UserService) mux.Middleware {
	return func(next mux.HandlerFunc) mux.HandlerFunc {
		return func(u *mux.UpdateCtx) error {
			from := u.Update.SentFrom()
			if from == nil {
				return next(u)
			}
			usr, err := users.GetByTelegramID(u.Ctx, from.ID)
			if err == nil && usr != nil {
				u.Ctx = context.WithValue(u.Ctx, userKey, usr)
			}
			return next(u)
		}
	}
}

func UserFrom(ctx context.Context) *models.User {
	v := ctx.Value(userKey)
	if v == nil {
		return nil
	}
	return v.(*models.User)
}
