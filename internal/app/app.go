package app

import (
	"context"
	"log"
	"os"
	"walki/config"
	"walki/internal/bot"
	"walki/internal/db"
	"walki/internal/repository/postgres"
	"walki/internal/service"
	"walki/internal/service/tgmedia"
	"walki/internal/storage/s3client"
)

// Run собирает зависимости и запускает бота
func Run() {
	cfg := config.LoadConfig()

	// init db pool
	connStr := os.Getenv("DATABASE_URL")
	pool := db.MustPool(connStr)
	defer pool.Close()

	s3c, err := s3client.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// repo-адаптеры поверх storage
	routeRepo := postgres.NewRouteRepo(pool)
	orderRepo := postgres.NewOrderRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	runRepo := postgres.NewRouteRunRepo(pool)
	mediaRepo := postgres.NewMediaRepo(pool)

	// сервисы
	routeSvc := service.NewRouteService(routeRepo)
	orderSvc := service.NewOrderService(orderRepo, routeRepo)
	profSvc := service.NewProfileService(orderRepo, routeRepo)
	userSvc := service.NewUserService(userRepo)
	runSvc := service.NewRouteRunService(routeRepo, orderRepo, runRepo)
	tgSvc := tgmedia.New(mediaRepo, s3c)

	// бот с явным внедрением сервисов
	b := bot.New(cfg.BotToken, routeSvc, orderSvc, profSvc, userSvc, runSvc, tgSvc)
	b.Start()
}
