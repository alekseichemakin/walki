package handlers

import (
	"log"
	"walki/internal/keyboards"
	"walki/internal/repository"
	"walki/internal/service"
	"walki/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler обрабатывает входящие сообщения и callback'и
type Handler struct {
	bot     *tgbotapi.BotAPI
	storage *storage.Storage
	// сервисы
	routes  *service.RouteService
	orders  *service.OrderService
	profile *service.ProfileService

	commands map[string]func(update tgbotapi.Update)
	buttons  map[string]func(update tgbotapi.Update)
}

// NewHandler создает новый экземпляр обработчика
func NewHandler(bot *tgbotapi.BotAPI, storage *storage.Storage) *Handler {
	h := &Handler{
		bot:      bot,
		storage:  storage,
		commands: make(map[string]func(update tgbotapi.Update)),
		buttons:  make(map[string]func(update tgbotapi.Update)),
	}

	// Transitional DI: поднимаем сервисы через адаптеры к storage
	routeRepo := repository.NewStorageRouteRepo(storage)
	orderRepo := repository.NewStorageOrderRepo(storage)

	h.routes = service.NewRouteService(routeRepo)
	h.orders = service.NewOrderService(orderRepo, routeRepo)
	h.profile = service.NewProfileService(orderRepo, routeRepo)

	h.registerCommands()
	h.registerButtons()

	return h
}

// RegisterCommand регистрирует обработчик команды
func (h *Handler) RegisterCommand(command string, handler func(update tgbotapi.Update)) {
	h.commands[command] = handler
}

// RegisterButton регистрирует обработчик кнопки
func (h *Handler) RegisterButton(button string, handler func(update tgbotapi.Update)) {
	h.buttons[button] = handler
}

// HandleUpdate обрабатывает входящее сообщение
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	// Обработка команд
	if update.Message.IsCommand() {
		if handler, ok := h.commands[update.Message.Command()]; ok {
			handler(update)
			return
		}
	}

	// Обработка кнопок
	if handler, ok := h.buttons[update.Message.Text]; ok {
		handler(update)
		return
	}

	// Сообщение по умолчанию
	h.sendMessage(update.Message.Chat.ID, "Выбери действие с клавиатуры ⌨️")
}

// sendMessage отправляет текстовое сообщение
func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// showMainMenu показывает главное меню
func (h *Handler) showMainMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Главное меню")
	msg.ReplyMarkup = keyboards.MainMenu()
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// Чистый конструктор с DI (используется из app/bot)
func NewHandlerWithServices(bot *tgbotapi.BotAPI, st *storage.Storage,
	rs *service.RouteService, os *service.OrderService, ps *service.ProfileService,
) *Handler {
	h := &Handler{
		bot:      bot,
		storage:  st,
		commands: make(map[string]func(update tgbotapi.Update)),
		buttons:  make(map[string]func(update tgbotapi.Update)),
		routes:   rs,
		orders:   os,
		profile:  ps,
	}
	h.registerCommands()
	h.registerButtons()
	return h
}
