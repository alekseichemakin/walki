package keyboards

import (
	"walki/internal/constants"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ButtonTexts содержит mapping констант на текст кнопок
var ButtonTexts = map[string]string{
	constants.BtnRoutes:  "🚶 Маршруты",
	constants.BtnProfile: "👤 Профиль",
	constants.BtnHelp:    "ℹ️ Помощь",
}

// MainMenu создает клавиатуру главного меню
func MainMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonTexts[constants.BtnRoutes]),
			tgbotapi.NewKeyboardButton(ButtonTexts[constants.BtnProfile]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonTexts[constants.BtnHelp]),
		),
	)
}

// MatchButton возвращает ключ кнопки по ее тексту
func MatchButton(text string) string {
	for key, val := range ButtonTexts {
		if val == text {
			return key
		}
	}
	return ""
}
