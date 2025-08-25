package handlers

import (
	"context"
	"fmt"
	"log"
	"walki/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handlePurchase(chatID int64, user *models.User, routeID int) {

	// Получаем информацию о маршруте для определения цены
	ctx := context.Background()
	route, err := h.routes.Details(ctx, routeID)
	if err != nil {
		log.Printf("Error getting route details for purchase: %v", err)
		h.sendMessage(chatID, "Ошибка при получении информации о маршруте")
		return
	}

	// Создаем заказ (заглушка оплаты)
	order, err := h.orders.Purchase(ctx, user.ID, routeID)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		h.sendMessage(chatID, "Ошибка при создании заказа")
		return
	}

	// Отправляем подтверждение покупки
	message := fmt.Sprintf(
		"🎉 Поздравляем с покупкой!\n\n"+
			"📍 *%s*\n"+
			"💰 Стоимость: %.2f руб.\n"+
			"📅 Доступен до: %s\n\n"+
			"Чтобы начать прогулку, перейдите в раздел \"👤 Профиль\" -> \"Мои маршруты\"",
		route.Title,
		order.Amount,
		order.AccessExpiry.Format("02.01.2006"),
	)

	// Кнопки для навигации
	profileBtn := tgbotapi.NewInlineKeyboardButtonData("👤 Перейти в профиль", CallbackMyRoutes)
	menuBtn := tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", CallbackMainMenu)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(profileBtn),
		tgbotapi.NewInlineKeyboardRow(menuBtn),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = markup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
