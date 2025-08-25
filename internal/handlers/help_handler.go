package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (h *Handler) handleHelp(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	h.bot.Send(tgbotapi.NewMessage(chatID, "ℹ️ Тут будет помощь"))

	// В будущем здесь будет:
	// 1. FAQ
	// 2. Инструкции
	// 3. Контакты поддержки
}
