package app

import (
	"os"
	"walki/config"
	"walki/internal/bot"
	"walki/internal/db"
	"walki/internal/repository/postgres"
	"walki/internal/service"
)

// Run собирает зависимости и запускает бота
func Run() {
	cfg := config.LoadConfig()

	// init db pool
	connStr := os.Getenv("DATABASE_URL")
	pool := db.MustPool(connStr)
	defer pool.Close()

	// repo-адаптеры поверх storage
	routeRepo := postgres.NewRouteRepo(pool)
	orderRepo := postgres.NewOrderRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	runRepo := postgres.NewRouteRunRepo(pool)

	// сервисы
	routeSvc := service.NewRouteService(routeRepo)
	orderSvc := service.NewOrderService(orderRepo, routeRepo)
	profSvc := service.NewProfileService(orderRepo, routeRepo)
	userSvc := service.NewUserService(userRepo)
	runSvc := service.NewRouteRunService(routeRepo, orderRepo, runRepo)

	// бот с явным внедрением сервисов
	b := bot.New(cfg.BotToken, routeSvc, orderSvc, profSvc, userSvc, runSvc)
	b.Start()
}
