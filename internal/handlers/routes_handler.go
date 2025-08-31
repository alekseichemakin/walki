package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleRoutes(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	h.showCitySelection(chatID)
}

func (h *Handler) showCitySelection(chatID int64) {
	cities, err := h.routes.Cities(context.Background())
	if err != nil {
		log.Printf("Error getting cities: %v", err)
		h.sendMessage(chatID, "Ошибка при загрузке городов")
		return
	}

	if len(cities) == 0 {
		h.sendMessage(chatID, "Пока нет маршрутов ни в одном городе")
		return
	}

	// Создаем инлайн-клавиатуру с городами
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, city := range cities {
		btn := tgbotapi.NewInlineKeyboardButtonData(city, CallbackCity+city)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// Кнопка "Назад" к главному меню
	backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", CallbackMainMenu)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backBtn))

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "Выберите город:")
	msg.ReplyMarkup = markup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) showRoutesByCity(chatID int64, city string) {
	routes, err := h.routes.ListByCity(context.Background(), city)
	if err != nil {
		log.Printf("Error getting routes for city %s: %v", city, err)
		h.sendMessage(chatID, "Ошибка при загрузке маршрутов")
		return
	}

	if len(routes) == 0 {
		h.sendMessage(chatID, fmt.Sprintf("В городе %s пока нет маршрутов", city))
		return
	}

	// Создаем инлайн-клавиатуру с маршрутами
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, route := range routes {
		btnText := fmt.Sprintf("📍 %s (%.1f км)", route.Title, route.LengthKm)
		btnData := CallbackRoute + strconv.Itoa(route.RouteID)
		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, btnData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// Кнопки навигации
	backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к городам", CallbackSelectCity)
	menuBtn := tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", CallbackMainMenu)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backBtn, menuBtn))

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Маршруты в городе %s:", city))
	msg.ReplyMarkup = markup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) showRouteDetails(chatID int64, routeID int) {
	version, err := h.routes.Details(context.Background(), routeID)
	if err != nil {
		log.Printf("Error getting route details for ID %d: %v", routeID, err)
		h.sendMessage(chatID, "Ошибка при загрузке информации о маршруте")
		return
	}

	message := fmt.Sprintf(
		"🚶 *%s*\n*Город:* %s\n*Описание:* %s\n*Протяженность:* %.1f км\n*Время прогулки:* %d мин\n*Тематика:* %s\n*Цена:* %.2f руб.",
		version.Title, version.City, version.Description, version.LengthKm,
		version.DurationMinutes, version.Theme, version.Price,
	)

	// Создаем кнопки для действий
	buyBtn := tgbotapi.NewInlineKeyboardButtonData("💰 Купить", CallbackBuy+strconv.Itoa(routeID))
	backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", CallbackSelectCity)
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buyBtn),
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)

	// Если есть обложка, отправляем фото с подписью
	if version.CoverImageURL != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(version.CoverImageURL))
		photo.Caption = message
		photo.ParseMode = "Markdown"
		photo.ReplyMarkup = markup

		if _, err := h.bot.Send(photo); err != nil {
			log.Printf("Error sending photo: %v", err)
			// Если не удалось отправить фото, отправляем текстовое сообщение
			h.sendMessageWithMarkup(chatID, message, markup)
		}
	} else {
		// Если нет обложки, отправляем текстовое сообщение
		h.sendMessageWithMarkup(chatID, message, markup)
	}
}

// Вспомогательная функция для отправки сообщения с разметкой
func (h *Handler) sendMessageWithMarkup(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = markup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
