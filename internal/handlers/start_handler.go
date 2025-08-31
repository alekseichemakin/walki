package handlers

import (
	"context"
	"log"
	"time"

	"walki/internal/keyboards"
	"walki/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleStart(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	user := update.Message.From

	// Создаем модель пользователя
	u := &models.User{
		TelegramID: user.ID,
		Username:   user.UserName,
		FullName:   user.FirstName + " " + user.LastName,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Сохраняем пользователя в БД
	if err := h.users.Register(context.Background(), u); err != nil {
		log.Printf("Ошибка сохранения пользователя: %v", err)
	}

	msg := tgbotapi.NewMessage(chatID, "Привет 👋 Добро пожаловать в Walki!")
	msg.ReplyMarkup = keyboards.MainMenu()
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
