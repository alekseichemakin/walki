package tgmedia

import (
	"context"
	"errors"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"walki/internal/repository"
	"walki/internal/storage/s3client"
)

type Service struct {
	Repo   repository.MediaRepository
	S3     *s3client.Client
	URLTTL time.Duration
}

func New(repo repository.MediaRepository, s3c *s3client.Client) *Service {
	return &Service{
		Repo:   repo,
		S3:     s3c,
		URLTTL: 5 * time.Minute,
	}
}

// SendMedia — единственная публичная точка.
// Определяет способ отправки по MIME: image/jpeg|png -> Photo, audio/* -> Audio, иначе -> Document.
// Возвращает сохранённый fileID и messageID отправленного сообщения.
func (s *Service) SendMedia(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	chatID int64,
	mediaID int64,
	caption string,
	parseMode string, // "", tgbotapi.ModeMarkdown(V2), tgbotapi.ModeHTML
) (fileID string, messageID int, err error) {

	// 1) Достаём метаданные
	m, err := s.Repo.GetByID(ctx, mediaID)
	if err != nil {
		return "", 0, err
	}
	if m == nil || m.S3Key == nil || *m.S3Key == "" {
		return "", 0, errors.New("media has no s3 object")
	}
	mime := ""
	if m.MimeType != nil {
		mime = *m.MimeType
	}

	kind := chooseKind(mime)

	// 2) Пробуем cached file_id
	if fid, ok, err := s.Repo.GetTelegramFileID(ctx, mediaID); err == nil && ok && fid != "" {
		switch kind {
		case "photo":
			msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fid))
			msg.Caption = caption
			if parseMode != "" {
				msg.ParseMode = parseMode
			}
			sent, err := bot.Send(msg)
			return fid, sent.MessageID, err

		case "audio":
			msg := tgbotapi.NewAudio(chatID, tgbotapi.FileID(fid))
			msg.Caption = caption
			if parseMode != "" {
				msg.ParseMode = parseMode
			}
			sent, err := bot.Send(msg)
			return fid, sent.MessageID, err

		default: // document
			msg := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fid))
			msg.Caption = caption
			if parseMode != "" {
				msg.ParseMode = parseMode
			}
			sent, err := bot.Send(msg)
			return fid, sent.MessageID, err
		}
	}

	// 3) Presigned URL из S3
	url, err := s.S3.PresignGet(ctx, *m.S3Key, s.URLTTL)
	if err != nil {
		return "", 0, err
	}

	// 4) Отправка по URL с кешированием вернувшегося file_id
	switch kind {
	case "photo":
		// Telegram как фото надёжно ест jpg/png; остальные форматы лучше документом.
		if !isPhotoMime(mime) {
			return s.sendDocumentURL(ctx, bot, chatID, mediaID, caption, parseMode, url, mime)
		}
		return s.sendPhotoURL(ctx, bot, chatID, mediaID, caption, parseMode, url, mime)

	case "audio":
		// Если это не audio/* — уходим в документ
		if !strings.HasPrefix(mime, "audio/") {
			return s.sendDocumentURL(ctx, bot, chatID, mediaID, caption, parseMode, url, mime)
		}
		return s.sendAudioURL(ctx, bot, chatID, mediaID, caption, parseMode, url, mime)

	default:
		return s.sendDocumentURL(ctx, bot, chatID, mediaID, caption, parseMode, url, mime)
	}
}

// -------------------- helpers (приватные) --------------------

func chooseKind(mime string) string {
	switch {
	case isPhotoMime(mime):
		return "photo"
	case strings.HasPrefix(mime, "audio/"):
		return "audio"
	default:
		return "document"
	}
}

func isPhotoMime(mime string) bool {
	m := strings.ToLower(mime)
	return strings.Contains(m, "jpeg") || strings.Contains(m, "jpg") || strings.Contains(m, "png")
}

func (s *Service) cacheTG(ctx context.Context, mediaID int64, fid, contentType string, chatID int64) {
	_ = s.Repo.UpsertTelegramFileID(ctx, mediaID, fid, contentType, &chatID)
}

func (s *Service) sendPhotoURL(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64,
	mediaID int64, caption, parseMode, url, mime string,
) (string, int, error) {

	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(url))
	msg.Caption = caption
	if parseMode != "" {
		msg.ParseMode = parseMode
	}
	sent, err := bot.Send(msg)
	if err != nil {
		return "", 0, err
	}
	if len(sent.Photo) == 0 {
		return "", 0, errors.New("telegram did not return photo file_id")
	}
	fid := sent.Photo[len(sent.Photo)-1].FileID
	s.cacheTG(ctx, mediaID, fid, mime, chatID)
	return fid, sent.MessageID, nil
}

func (s *Service) sendAudioURL(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64,
	mediaID int64, caption, parseMode, url, mime string,
) (string, int, error) {

	msg := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(url))
	msg.Caption = caption
	if parseMode != "" {
		msg.ParseMode = parseMode
	}
	sent, err := bot.Send(msg)
	if err != nil {
		return "", 0, err
	}
	if sent.Audio == nil {
		return "", 0, errors.New("telegram did not return audio file_id")
	}
	fid := sent.Audio.FileID
	s.cacheTG(ctx, mediaID, fid, mime, chatID)
	return fid, sent.MessageID, nil
}

func (s *Service) sendDocumentURL(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64,
	mediaID int64, caption, parseMode, url, mime string,
) (string, int, error) {

	msg := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(url))
	msg.Caption = caption
	if parseMode != "" {
		msg.ParseMode = parseMode
	}
	sent, err := bot.Send(msg)
	if err != nil {
		return "", 0, err
	}
	if sent.Document == nil {
		return "", 0, errors.New("telegram did not return document file_id")
	}
	fid := sent.Document.FileID
	s.cacheTG(ctx, mediaID, fid, mime, chatID)
	return fid, sent.MessageID, nil
}
