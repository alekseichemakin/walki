package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"walki/internal/handlers"
	"walki/internal/service"
)

// Bot представляет собой экземпляр Telegram бота
type Bot struct {
	api     *tgbotapi.BotAPI
	handler *handlers.Handler
}

// Новый конструктор с DI сервисов.
func New(token string,
	routeSvc *service.RouteService,
	orderSvc *service.OrderService,
	profileSvc *service.ProfileService,
	userSvc *service.UserService,
	runSvc *service.RouteRunService) *Bot {
	api, _ := tgbotapi.NewBotAPI(token)
	h := handlers.NewHandler(api, routeSvc, orderSvc, profileSvc, userSvc, runSvc)
	return &Bot{api: api, handler: h}
}

// Start запускает бота в режиме опроса
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			// Обрабатываем callback-запросы
			b.handler.HandleCallback(update)
		} else if update.Message != nil {
			// Обрабатываем обычные сообщения
			b.handler.HandleUpdate(update)
		}
	}
}
