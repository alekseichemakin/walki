package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleProfile(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	// Создаем инлайн-клавиатуру для профиля
	myRoutesBtn := tgbotapi.NewInlineKeyboardButtonData("🚶 Мои маршруты", CallbackMyRoutes)
	backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", CallbackMainMenu)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(myRoutesBtn),
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)

	msg := tgbotapi.NewMessage(chatID, "👤 Ваш профиль\n\nЗдесь вы можете управлять своими маршрутами и настройки")
	msg.ReplyMarkup = markup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) showUserRoutes(chatID int64, userID int) {
	// Получаем маршруты пользователя
	orders, err := h.profile.MyOrders(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting user orders: %v", err)
		h.sendMessage(chatID, "Ошибка при загрузке ваших маршрутов")
		return
	}

	if len(orders) == 0 {
		h.sendMessage(chatID, "У вас пока нет купленных маршрутов")
		return
	}

	// Создаем кнопки для каждого маршрута
	var routeButtons []tgbotapi.InlineKeyboardButton
	for _, order := range orders {
		btnText := fmt.Sprintf("📍 %s (%s)", order.RouteTitle, order.RouteCity)
		btnData := CallbackPurchasedRoute + strconv.Itoa(order.RouteID)
		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, btnData)
		routeButtons = append(routeButtons, btn)
	}

	// Создаем rows для клавиатуры (максимум 1 кнопка в ряду)
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, btn := range routeButtons {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// Добавляем кнопку "Назад"
	backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", CallbackMainMenu)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backBtn))

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, "🚶 *Ваши маршруты:*\n\nВыберите маршрут для просмотра деталей:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = markup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) showPurchasedRouteDetails(chatID int64, userID int, routeID int) {
	// Проверяем, есть ли у пользователя доступ к этому маршруту
	hasAccess, err := h.profile.HasAccess(context.Background(), userID, routeID)
	if err != nil {
		log.Printf("Error checking access: %v", err)
		h.sendMessage(chatID, "Ошибка при проверке доступа к маршруту")
		return
	}

	if !hasAccess {
		h.sendMessage(chatID, "У вас нет доступа к этому маршруту")
		return
	}

	// Получаем информацию о заказе для получения даты истечения доступа
	orders, err := h.profile.MyOrders(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting user orders: %v", err)
		h.sendMessage(chatID, "Ошибка при загрузке информации о доступе")
		return
	}

	// Ищем заказ и версию для этого маршрута
	var message string
	for _, order := range orders {
		if order.RouteID == routeID {
			route, err := h.routes.VersionByID(context.Background(), order.VersionID)
			if err != nil {
				log.Printf("Error getting route version: %v", err)
				h.sendMessage(chatID, "Ошибка при загрузке информации о версии маршрута")
				return
			}
			var expiryInfo string
			if order.AccessExpiry != nil {
				expiryInfo = fmt.Sprintf("доступен до %s", order.AccessExpiry.Format("02.01.2006"))
			} else {
				expiryInfo = "бессрочный доступ"
			}
			// Формируем сообщение с информацией о маршруте
			message = fmt.Sprintf(
				"🚶 *%s*\n*Город:* %s\n*Описание:* %s\n*Протяженность:* %.1f км\n*Время прогулки:* %d мин\n\n*Статус:* %s",
				route.Title, route.City, route.Description, route.LengthKm, route.DurationMinutes, expiryInfo,
			)
			break
		}
	}

	// Создаем кнопки для управления маршрутом
	startRouteBtn := tgbotapi.NewInlineKeyboardButtonData(
		"🎯 Начать прогулку",
		CallbackStartRoute+strconv.Itoa(routeID),
	)
	backBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", CallbackMyRoutes)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(startRouteBtn),
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = markup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
