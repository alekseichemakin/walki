package handlers

import (
	"log"
	"strconv"
	"strings"
	"walki/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Callback types
const (
	CallbackCity           = "city:"
	CallbackRoute          = "route:"
	CallbackBuy            = "buy:"
	CallbackSelectCity     = "action:select_city"
	CallbackMainMenu       = "menu:main"
	CallbackMyRoutes       = "profile:my_routes"
	CallbackStartRoute     = "routes:start"
	CallbackPurchasedRoute = "purchased_route:"
)

// CallbackHandler представляет функцию для обработки callback'а
type CallbackHandler func(chatID int64, user *models.User, data string)

func (h *Handler) HandleCallback(update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}

	callback := update.CallbackQuery
	data := callback.Data
	chatID := callback.Message.Chat.ID
	userTgID := callback.From.ID

	// Получаем пользователя
	user, err := h.storage.GetUserByTelegramID(userTgID)
	if err != nil {
		log.Printf("User not found by tgID: %d", userTgID)
		h.sendMessage(chatID, "Ошибка: Пользователь не найден попробуйте начать с команды /start")
		return
	}

	// Обрабатываем callback
	h.routeCallback(data, chatID, user)

	// Ответим на callback, чтобы убрать "часики" у кнопки
	h.answerCallback(callback.ID)
}

// routeCallback определяет тип callback и направляет его соответствующему обработчику
func (h *Handler) routeCallback(data string, chatID int64, user *models.User) {
	// Создаем карту обработчиков для точных совпадений
	exactHandlers := map[string]CallbackHandler{
		CallbackSelectCity: h.handleSelectCity,
		CallbackMainMenu:   h.handleMainMenu,
		CallbackMyRoutes:   h.handleMyRoutes,
	}

	// Проверяем точные совпадения
	if handler, exists := exactHandlers[data]; exists {
		handler(chatID, user, data)
		return
	}

	// Создаем карту обработчиков для префиксных callback'ов
	prefixHandlers := map[string]CallbackHandler{
		CallbackCity:           h.handleCityCallback,
		CallbackRoute:          h.handleRouteCallback,
		CallbackBuy:            h.handleBuyCallback,
		CallbackPurchasedRoute: h.handlePurchasedRouteCallback,
		CallbackStartRoute:     h.handleStartRouteCallback,
	}

	// Проверяем префиксные callback'и
	for prefix, handler := range prefixHandlers {
		if strings.HasPrefix(data, prefix) {
			handler(chatID, user, data)
			return
		}
	}

	// Если не нашли подходящий обработчик
	h.sendMessage(chatID, "Неизвестная команда")
}

// Обработчики для каждого типа callback'а

func (h *Handler) handleCityCallback(chatID int64, user *models.User, data string) {
	city := strings.TrimPrefix(data, CallbackCity)
	h.showRoutesByCity(chatID, city)
}

func (h *Handler) handleRouteCallback(chatID int64, user *models.User, data string) {
	routeIDStr := strings.TrimPrefix(data, CallbackRoute)
	routeID, err := strconv.Atoi(routeIDStr)
	if err != nil {
		log.Printf("Invalid route ID: %s", routeIDStr)
		h.sendMessage(chatID, "Ошибка: неверный идентификатор маршрута")
		return
	}
	h.showRouteDetails(chatID, routeID)
}

func (h *Handler) handleBuyCallback(chatID int64, user *models.User, data string) {
	routeIDStr := strings.TrimPrefix(data, CallbackBuy)
	routeID, err := strconv.Atoi(routeIDStr)
	if err != nil {
		log.Printf("Invalid route ID for purchase: %s", routeIDStr)
		h.sendMessage(chatID, "Ошибка: неверный идентификатор маршрута")
		return
	}
	h.handlePurchase(chatID, user, routeID)
}

func (h *Handler) handlePurchasedRouteCallback(chatID int64, user *models.User, data string) {
	routeIDStr := strings.TrimPrefix(data, CallbackPurchasedRoute)
	routeID, err := strconv.Atoi(routeIDStr)
	if err != nil {
		log.Printf("Invalid purchased route ID: %s", routeIDStr)
		h.sendMessage(chatID, "Ошибка: неверный идентификатор маршрута")
		return
	}
	h.showPurchasedRouteDetails(chatID, user.ID, routeID)
}

func (h *Handler) handleSelectCity(chatID int64, user *models.User, data string) {
	h.showCitySelection(chatID)
}

func (h *Handler) handleMainMenu(chatID int64, user *models.User, data string) {
	h.showMainMenu(chatID)
}

func (h *Handler) handleMyRoutes(chatID int64, user *models.User, data string) {
	h.showUserRoutes(chatID, user.ID)
}

func (h *Handler) handleStartRouteCallback(chatID int64, user *models.User, data string) {
	h.sendMessage(chatID, "Функция начала маршрута скоро будет доступна!")
}

// answerCallback отвечает на callback запрос
func (h *Handler) answerCallback(callbackID string) {
	callbackConf := tgbotapi.NewCallback(callbackID, "")
	if _, err := h.bot.Request(callbackConf); err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}
