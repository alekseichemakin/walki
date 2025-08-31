package service

import (
	"context"
	"errors"
	"fmt"

	"walki/internal/models"
	"walki/internal/repository"
)

// Доменно-значимые ошибки
var (
	ErrNoAccess = errors.New("user has no access to the route")
)

type RouteRunService struct {
	routes  repository.RouteRepository
	orders  repository.OrderRepository
	runRepo repository.RouteRunRepository
}

func NewRouteRunService(rr repository.RouteRepository, or repository.OrderRepository, run repository.RouteRunRepository) *RouteRunService {
	return &RouteRunService{routes: rr, orders: or, runRepo: run}
}

// DTO, который отдаём телеграм-слою
type PointWithMedia struct {
	Point        *models.RoutePoint
	Photos       []string
	Voices       []string
	VersionID    int
	RouteID      int
	Idx          int
	HasPrev      bool
	HasNext      bool
	ContentMsgID *int // message_id «контент»-сообщения (фото/текст)
	VoiceMsgID   *int // message_id voice-сообщения
}

/* ==========================
   Публичные методы сервиса
   ========================== */

// Старт: проверяем доступ и переходим к первой точке версии маршрута
func (s *RouteRunService) Start(ctx context.Context, userID, routeID int) (*PointWithMedia, error) {
	if err := s.ensureAccess(ctx, userID, routeID); err != nil {
		return nil, err
	}
	return s.moveFirst(ctx, userID, routeID)
}

// Текущий прогресс (nil, nil если не найден — это поведение часто удобно наверху)
func (s *RouteRunService) Progress(ctx context.Context, userID, routeID int) (*models.RouteProgress, error) {
	ver, err := s.getVersion(ctx, routeID)
	if err != nil {
		return nil, err
	}
	pr, err := s.runRepo.GetProgress(ctx, userID, ver.ID)
	if err != nil {
		// Отсутствие прогресса не считаем ошибкой домена — возвращаем (nil, nil),
		// чтобы вызывающий слой мог сам решить, что делать (например, предложить Start).
		return nil, nil
	}
	return pr, nil
}

// Continue: шаг вперёд от текущего индекса; если дальше нет точек — завершаем прогресс.
func (s *RouteRunService) Continue(ctx context.Context, userID, routeID int) (*PointWithMedia, bool, error) {
	ver, err := s.getVersion(ctx, routeID)
	if err != nil {
		return nil, false, err
	}
	pr, err := s.runRepo.GetProgress(ctx, userID, ver.ID)
	if err != nil || pr == nil {
		return nil, false, nil // нет прогресса — пусть UI решит, звать Start или показать подсказку
	}

	nextIdx, hasNext, err := s.runRepo.NextIndex(ctx, ver.ID, pr.CurrentIdx)
	if err != nil {
		return nil, false, err
	}
	if !hasNext {
		// достигли конца — фиксируем завершение
		if finishErr := s.runRepo.Finish(ctx, userID, ver.ID); finishErr != nil {
			// логически это не «фатал», но ошибку лучше не терять
			return nil, false, fmt.Errorf("finish route: %w", finishErr)
		}
		return nil, false, nil
	}
	pm, err := s.moveToIndex(ctx, userID, ver, nextIdx)
	if err != nil {
		return nil, false, err
	}
	return pm, true, nil
}

// Restart: сброс на первую точку
func (s *RouteRunService) Restart(ctx context.Context, userID, routeID int) (*PointWithMedia, error) {
	return s.moveFirst(ctx, userID, routeID)
}

// Prev: шаг назад от текущего индекса
func (s *RouteRunService) Prev(ctx context.Context, userID, routeID int) (*PointWithMedia, bool, error) {
	ver, err := s.getVersion(ctx, routeID)
	if err != nil {
		return nil, false, err
	}
	pr, err := s.runRepo.GetProgress(ctx, userID, ver.ID)
	if err != nil || pr == nil {
		return nil, false, nil
	}
	prevIdx, hasPrev, err := s.runRepo.PrevIndex(ctx, ver.ID, pr.CurrentIdx)
	if err != nil {
		return nil, false, err
	}
	if !hasPrev {
		return nil, false, nil
	}
	pm, err := s.moveToIndex(ctx, userID, ver, prevIdx)
	if err != nil {
		return nil, false, err
	}
	return pm, true, nil
}

// Явное завершение
func (s *RouteRunService) Finish(ctx context.Context, userID, routeVerID int) error {
	return s.runRepo.Finish(ctx, userID, routeVerID)
}

func (s *RouteRunService) UpdateMessageIDs(ctx context.Context, userID, versionID int, contentMsgID, voiceMsgID *int) error {
	return s.runRepo.UpdateMessageIDs(ctx, userID, versionID, contentMsgID, voiceMsgID)
}

/* ==========================
   Внутренние хелперы
   ========================== */

// Проверка доступа пользователя к роуту
func (s *RouteRunService) ensureAccess(ctx context.Context, userID, routeID int) error {
	ok, err := s.orders.UserHasAccess(ctx, userID, routeID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoAccess
	}
	return nil
}

// Детали активной версии по routeID
func (s *RouteRunService) getVersion(ctx context.Context, routeID int) (*models.RouteVersion, error) {
	ver, err := s.routes.Details(ctx, routeID)
	if err != nil {
		return nil, err
	}
	return ver, nil
}

// Перейти к первой точке версии и заапсертить прогресс
func (s *RouteRunService) moveFirst(ctx context.Context, userID, routeID int) (*PointWithMedia, error) {
	ver, err := s.getVersion(ctx, routeID)
	if err != nil {
		return nil, err
	}
	// попробуем явный idx=0, если его нет — возьмём FirstPoint
	p, err := s.runRepo.PointByIndex(ctx, ver.ID, 0)
	if err != nil {
		if p, err = s.runRepo.FirstPoint(ctx, ver.ID); err != nil {
			return nil, err
		}
	}
	return s.moveToIndex(ctx, userID, ver, p.Idx)
}

// Универсальный переход к произвольному индексу (апдейт прогресса + упаковка ответа)
func (s *RouteRunService) moveToIndex(ctx context.Context, userID int, ver *models.RouteVersion, idx int) (*PointWithMedia, error) {
	p, err := s.runRepo.PointByIndex(ctx, ver.ID, idx)
	if err != nil {
		return nil, err
	}
	if err := s.runRepo.UpsertProgress(ctx, userID, ver.RouteID, ver.ID, p.Idx); err != nil {
		return nil, err
	}
	return s.pack(ctx, userID, ver.ID, ver.RouteID, p)
}

// Есть ли соседние точки (без превращения ошибок в ложные флаги)
func (s *RouteRunService) hasAdj(ctx context.Context, versionID, idx int) (bool, bool) {
	_, hasPrev, errPrev := s.runRepo.PrevIndex(ctx, versionID, idx)
	_, hasNext, errNext := s.runRepo.NextIndex(ctx, versionID, idx)
	// если ошибки, считаем, что соответствующего соседа нет (и не паникуем)
	if errPrev != nil {
		hasPrev = false
	}
	if errNext != nil {
		hasNext = false
	}
	return hasPrev, hasNext
}

// Сборка ответа для телеги: медиа, флаги соседей и message_id из прогресса
func (s *RouteRunService) pack(
	ctx context.Context,
	userID int,
	verID int,
	routeID int,
	p *models.RoutePoint,
) (*PointWithMedia, error) {

	photos, voices, err := s.runRepo.PointMedia(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	hasPrev, hasNext := s.hasAdj(ctx, verID, p.Idx)

	// подгрузим прогресс, чтобы вернуть msgID — это важно для edit/replace в хендлере
	pr, _ := s.runRepo.GetProgress(ctx, userID, verID)

	var contentID, voiceID *int
	if pr != nil {
		if pr.ContentMsgID != nil {
			cid := *pr.ContentMsgID
			contentID = &cid
		}
		if pr.VoiceMsgID != nil {
			vid := *pr.VoiceMsgID
			voiceID = &vid
		}
	}

	return &PointWithMedia{
		Point:        p,
		Photos:       photos,
		Voices:       voices,
		VersionID:    verID,
		RouteID:      routeID,
		Idx:          p.Idx,
		HasPrev:      hasPrev,
		HasNext:      hasNext,
		ContentMsgID: contentID,
		VoiceMsgID:   voiceID,
	}, nil
}
