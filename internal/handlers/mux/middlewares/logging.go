package middlewares

import (
	"log"

	"walki/internal/handlers/mux"
)

func Logging() mux.Middleware {
	return func(next mux.HandlerFunc) mux.HandlerFunc {
		return func(u *mux.UpdateCtx) error {
			log.Printf("[update] chat=%d cb=%v cmd=%v",
				u.ChatID,
				u.Update.CallbackQuery != nil,
				u.Update.Message != nil && u.Update.Message.IsCommand(),
			)
			return next(u)
		}
	}
}
