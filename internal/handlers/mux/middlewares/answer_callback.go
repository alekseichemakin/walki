package middlewares

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"walki/internal/handlers/mux"
)

// Авто-ACK для инлайн-кнопок, чтобы у кнопки пропадали «часики».
func AnswerCallback() mux.Middleware {
	return func(next mux.HandlerFunc) mux.HandlerFunc {
		return func(u *mux.UpdateCtx) error {
			if u.Update.CallbackQuery != nil {
				defer func() {
					_, err := u.Sender.Request(tgbotapi.NewCallback(u.Update.CallbackQuery.ID, ""))
					if err != nil {
						log.Printf("answerCallback error: %v", err)
					}
				}()
			}
			return next(u)
		}
	}
}
