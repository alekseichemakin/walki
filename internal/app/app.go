package app

import (
	"walki/config"
	"walki/internal/bot"
	"walki/internal/repository"
	"walki/internal/service"
	"walki/internal/storage"
)

// Run собирает зависимости и запускает бота
func Run() {
	cfg := config.LoadConfig()

	// текущее хранилище (как и раньше)
	st := storage.NewStorage()
	defer st.Close()

	// repo-адаптеры поверх storage
	routeRepo := repository.NewStorageRouteRepo(st)
	orderRepo := repository.NewStorageOrderRepo(st)
	userRepo := repository.NewStorageUserRepo(st)

	// сервисы
	routeSvc := service.NewRouteService(routeRepo)
	orderSvc := service.NewOrderService(orderRepo, routeRepo)
	profSvc := service.NewProfileService(orderRepo, routeRepo)
	_ = userRepo // пригодится в следующих шагах (start/онбординг и т.п.)

	// бот с явным внедрением сервисов
	b := bot.New(cfg.BotToken, st, routeSvc, orderSvc, profSvc)
	b.Start()
}
