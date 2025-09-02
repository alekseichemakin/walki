package handlers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"walki/internal/service"
)

/*
Прохождение точки маршрута в два шага:
1) Контент (фото + caption, либо текст)
2) Голос (отдельным сообщением)
Всегда удаляем предыдущие сообщения, чтобы карточка была внизу чата.
*/
func (h *Handler) renderRoutePoint(chatID int64, userID int, data *service.PointWithMedia) {
	kb := h.navKeyboard(data.RouteID, data.HasPrev, data.HasNext)
	caption := buildCaption(data)

	// Поднимаем карточку: удаляем старые content+voice, шлём новые
	h.deleteIfExists(chatID, data.ContentMsgID)
	h.deleteIfExists(chatID, data.VoiceMsgID)

	if err := h.sendFreshContent(chatID, userID, data, caption, kb); err != nil {
		log.Printf("send fresh content: %v", err)
	}
	if err := h.sendFreshVoice(chatID, userID, data); err != nil {
		log.Printf("send fresh voice: %v", err)
	}
}

// Удаление сообщения по id (если оно есть)
func (h *Handler) deleteIfExists(chatID int64, msgID *int) {
	if msgID == nil || *msgID == 0 {
		return
	}
	del := tgbotapi.NewDeleteMessage(chatID, *msgID)
	if _, err := h.bot.Request(del); err != nil {
		log.Printf("delete message %d: %v", *msgID, err)
	}
}

// Отправить новое контент-сообщение (фото+подпись или текст) и сохранить его message_id
func (h *Handler) sendFreshContent(chatID int64, userID int, data *service.PointWithMedia, caption string, kb tgbotapi.InlineKeyboardMarkup) error {
	// если есть фото — отправляем его через сервис (кэш TG + presigned S3)
	if len(data.PhotoIds) > 0 {
		_, msgID, err := h.tgMedia.SendMedia(h.ctx(), h.bot, chatID, data.PhotoIds[0], caption, "Markdown")
		if err != nil {
			return err
		}
		return h.run.UpdateMessageIDs(h.ctx(), userID, data.VersionID, &msgID, nil)
	}

	// иначе — текстовая «страница»
	m := tgbotapi.NewMessage(chatID, caption)
	m.ParseMode = "Markdown"
	m.ReplyMarkup = kb
	sent, err := h.bot.Send(m)
	if err != nil {
		return err
	}
	return h.run.UpdateMessageIDs(h.ctx(), userID, data.VersionID, &sent.MessageID, nil)
}

// Отправить новое voice-/audio-сообщение (если есть) и сохранить его message_id, иначе очистить voice_msg_id
func (h *Handler) sendFreshVoice(chatID int64, userID int, data *service.PointWithMedia) error {
	if len(data.VoiceIds) == 0 {
		// очистить voice в прогрессе
		zero := 0 // репозиторий должен трактовать 0 как NULL (через NULLIF)
		return h.run.UpdateMessageIDs(h.ctx(), userID, data.VersionID, nil, &zero)
	}

	// Отправляем через сервис (сам решит, чем слать по MIME; для аудио это будет Audio/Document)
	_, msgID, err := h.tgMedia.SendMedia(h.ctx(), h.bot, chatID, data.VoiceIds[0], "", "")
	if err != nil {
		return err
	}
	return h.run.UpdateMessageIDs(h.ctx(), userID, data.VersionID, nil, &msgID)
}

/* =========================
   UI/вёрстка
   ========================= */

func (h *Handler) navKeyboard(routeID int, hasPrev, hasNext bool) tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	if hasPrev {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "route_prev:"+strconv.Itoa(routeID)))
	}
	if hasNext {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️ Дальше", "route_next:"+strconv.Itoa(routeID)))
	} else {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("🏁 Завершить", "route_finish:"+strconv.Itoa(routeID)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(row...))
}

func buildCaption(p *service.PointWithMedia) string {
	title := "📍 *" + escapeMd(p.Point.Title) + "*"
	desc := trimTo(850, escapeMd(p.Point.Description)) // чтобы вместе с координатами уложиться в лимит
	lat := fmt.Sprintf("%.6f", p.Point.Lat)
	lon := fmt.Sprintf("%.6f", p.Point.Lon)
	mapsURL := "https://maps.google.com/?q=" + url.QueryEscape(lat+","+lon)

	// без Markdown-ссылок — просто URL, меньше шансов словить parse error
	return fmt.Sprintf("%s\n\n%s\n\n`%s, %s`\n%s", title, desc, lat, lon, mapsURL)
}

func escapeMd(s string) string {
	return strings.NewReplacer("_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "`", "\\`").Replace(s)
}

func trimTo(n int, s string) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n-1]) + "…"
}

func (h *Handler) ctx() context.Context { return context.Background() }
