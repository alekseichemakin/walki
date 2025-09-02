package handlers

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"walki/internal/constants"
	"walki/internal/handlers/mux"
	"walki/internal/handlers/mux/middlewares"
	"walki/internal/keyboards"
	"walki/internal/service"
	"walki/internal/service/tgmedia"
)

// Handler обрабатывает входящие сообщения и callback'и
type Handler struct {
	bot *tgbotapi.BotAPI
	// сервисы
	routes  *service.RouteService
	orders  *service.OrderService
	profile *service.ProfileService
	users   *service.UserService
	run     *service.RouteRunService
	tgMedia *tgmedia.Service
	router  *mux.Router
}

// Чистый конструктор с DI (используется из app/bot)
func NewHandler(bot *tgbotapi.BotAPI,
	rs *service.RouteService,
	os *service.OrderService,
	ps *service.ProfileService,
	userSvc *service.UserService,
	runSvc *service.RouteRunService,
	tg *tgmedia.Service) *Handler {
	h := &Handler{
		bot:     bot,
		routes:  rs,
		orders:  os,
		profile: ps,
		users:   userSvc,
		run:     runSvc,
		tgMedia: tg,
	}

	// --- Router  middlewares
	r := mux.New()
	r.Use(middlewares.Logging())
	r.Use(middlewares.AnswerCallback())
	r.Use(middlewares.WithUser(h.users))
	//r.Use(middlewares.Timeout(5 * time.Second)) // при желании

	// === Команды
	r.Command("start", func(u *mux.UpdateCtx) error { h.handleStart(u.Update); return nil })
	r.Command("routes", func(u *mux.UpdateCtx) error { h.handleRoutes(u.Update); return nil })
	r.Command("profile", func(u *mux.UpdateCtx) error { h.handleProfile(u.Update); return nil })
	r.Command("help", func(u *mux.UpdateCtx) error { h.handleHelp(u.Update); return nil })

	// === Кнопки главного меню (точный текст)
	r.Message(keyboards.ButtonTexts[constants.BtnRoutes], func(u *mux.UpdateCtx) error { h.handleRoutes(u.Update); return nil })
	r.Message(keyboards.ButtonTexts[constants.BtnProfile], func(u *mux.UpdateCtx) error { h.handleProfile(u.Update); return nil })
	r.Message(keyboards.ButtonTexts[constants.BtnHelp], func(u *mux.UpdateCtx) error { h.handleHelp(u.Update); return nil })

	// === Callback’и (точные)
	r.CallbackExact("action:select_city", func(u *mux.UpdateCtx) error { h.showCitySelection(u.ChatID); return nil })
	r.CallbackExact("menu:main", func(u *mux.UpdateCtx) error { h.showMainMenu(u.ChatID); return nil })
	r.CallbackExact("profile:my_routes", func(u *mux.UpdateCtx) error {
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}
		h.showUserRoutes(u.ChatID, usr.ID)
		return nil
	})

	// === Callback’и (префиксы)
	r.CallbackPrefix(CallbackCity, func(u *mux.UpdateCtx, v mux.Values) error {
		h.showRoutesByCity(u.ChatID, v["id"])
		return nil
	})
	r.CallbackPrefix(CallbackRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		id, _ := strconv.Atoi(v["id"])
		h.showRouteDetails(u.ChatID, id)
		return nil
	})
	r.CallbackPrefix(CallbackBuy, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}
		h.handlePurchase(u.ChatID, usr, routeID)
		return nil
	})
	r.CallbackPrefix(CallbackPurchasedRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}
		h.showPurchasedRouteDetails(u.ChatID, usr.ID, routeID)
		return nil
	})
	// в NewHandler, там где регистрируешь router r := mux.New()
	r.CallbackPrefix(CallbackStartRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}

		// есть ли незавершённый прогресс?
		if pr, _ := h.run.Progress(u.Ctx, usr.ID, routeID); pr != nil && pr.FinishedAt == nil {
			// спросим: продолжить или начать заново
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("▶️ Продолжить", fmt.Sprintf("route_continue:%d", routeID)),
					tgbotapi.NewInlineKeyboardButtonData("🔁 Начать заново", fmt.Sprintf("route_restart:%d", routeID)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "menu:main"),
				),
			)
			msg := tgbotapi.NewMessage(u.ChatID, "У вас уже начат маршрут. Что сделать?")
			msg.ReplyMarkup = kb
			_, _ = h.bot.Send(msg)
			return nil
		}

		// иначе — стартуем
		res, err := h.run.Start(u.Ctx, usr.ID, routeID)
		if err != nil {
			h.sendMessage(u.ChatID, "Нет доступа или маршрут пуст.")
			return nil
		}
		h.renderRoutePoint(u.ChatID, usr.ID, res)
		return nil
	})

	r.CallbackPrefix(CallbackContinueRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}

		res, ok, err := h.run.Continue(u.Ctx, usr.ID, routeID)
		if err != nil {
			h.sendMessage(u.ChatID, "Произошла ошибка.")
			return nil
		}
		if !ok {
			h.sendMessage(u.ChatID, "🏁 Маршрут уже завершён.")
			return nil
		}
		h.renderRoutePoint(u.ChatID, usr.ID, res)
		return nil
	})

	r.CallbackPrefix(CallbackRestartRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}

		res, err := h.run.Restart(u.Ctx, usr.ID, routeID)
		if err != nil {
			h.sendMessage(u.ChatID, "Не получилось начать заново.")
			return nil
		}
		h.renderRoutePoint(u.ChatID, usr.ID, res)
		return nil
	})

	r.CallbackPrefix(CallbackNextRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}

		res, ok, err := h.run.Continue(u.Ctx, usr.ID, routeID)
		if err != nil {
			h.sendMessage(u.ChatID, "Ошибка, попробуйте ещё раз.")
			return nil
		}
		if !ok {
			h.sendMessage(u.ChatID, "🏁 Маршрут завершён! Возвращайтесь в профиль, чтобы пройти ещё раз.")
			return nil
		}
		h.renderRoutePoint(u.ChatID, usr.ID, res)
		return nil
	})

	r.CallbackPrefix(CallbackPrevRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}

		res, ok, err := h.run.Prev(u.Ctx, usr.ID, routeID)
		if err != nil {
			h.sendMessage(u.ChatID, "Ошибка, попробуйте ещё раз.")
			return nil
		}
		if !ok {
			h.sendMessage(u.ChatID, "Это первая точка маршрута.")
			return nil
		}
		h.renderRoutePoint(u.ChatID, usr.ID, res)
		return nil
	})

	r.CallbackPrefix(CallbackFinishRoute, func(u *mux.UpdateCtx, v mux.Values) error {
		routeID, _ := strconv.Atoi(v["id"])
		usr := middlewares.UserFrom(u.Ctx)
		if usr == nil {
			h.sendMessage(u.ChatID, "Ошибка: пользователь не найден, попробуйте /start")
			return nil
		}

		ver, _ := h.routes.Details(u.Ctx, routeID)
		if ver != nil {
			err := h.run.Finish(u.Ctx, usr.ID, ver.ID)
			if err != nil {
				h.sendMessage(u.ChatID, "Ошибка: не удалось завершить маршрут, обратитесь в поддержку")
				return nil
			}
		}
		h.sendMessage(u.ChatID, "🏁 Маршрут завершён! Спасибо за прогулку.")
		return nil
	})

	// Дефолт
	r.Default(func(u *mux.UpdateCtx) error {
		_, _ = h.bot.Send(tgbotapi.NewMessage(u.ChatID, "Выбери действие с клавиатуры ⌨️"))
		return nil
	})

	h.router = r

	return h
}

// HandleUpdate обрабатывает входящее сообщение
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	_ = h.router.Dispatch(&mux.UpdateCtx{
		Ctx:    context.Background(),
		Update: update,
		ChatID: update.Message.Chat.ID,
		Sender: h.bot,
	})
}

func (h *Handler) HandleCallback(update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}
	_ = h.router.Dispatch(&mux.UpdateCtx{
		Ctx:    context.Background(),
		Update: update,
		ChatID: update.CallbackQuery.Message.Chat.ID,
		Sender: h.bot,
	})
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
