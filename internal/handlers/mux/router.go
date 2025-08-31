package mux

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

type UpdateCtx struct {
	Ctx    context.Context
	Update tgbotapi.Update
	ChatID int64
	Sender Sender
}

type HandlerFunc func(*UpdateCtx) error
type CallbackFunc func(*UpdateCtx, Values) error
type Middleware func(HandlerFunc) HandlerFunc

type Router struct {
	mw       []Middleware
	commands map[string]HandlerFunc
	messages map[string]HandlerFunc
	cbExact  map[string]HandlerFunc
	cbPrefix map[string]CallbackFunc
	def      HandlerFunc
}

func New() *Router {
	return &Router{
		commands: map[string]HandlerFunc{},
		messages: map[string]HandlerFunc{},
		cbExact:  map[string]HandlerFunc{},
		cbPrefix: map[string]CallbackFunc{},
	}
}

func (r *Router) Use(m Middleware) { r.mw = append(r.mw, m) }

func (r *Router) Command(cmd string, h HandlerFunc)            { r.commands[cmd] = h }
func (r *Router) Message(text string, h HandlerFunc)           { r.messages[text] = h }
func (r *Router) CallbackExact(data string, h HandlerFunc)     { r.cbExact[data] = h }
func (r *Router) CallbackPrefix(prefix string, h CallbackFunc) { r.cbPrefix[prefix] = h }
func (r *Router) Default(h HandlerFunc)                        { r.def = h }

func (r *Router) Dispatch(u *UpdateCtx) bool {
	h, ok := r.pick(u)
	if !ok {
		return false
	}
	for i := len(r.mw) - 1; i >= 0; i-- {
		h = r.mw[i](h)
	}
	_ = h(u) // ошибки можно централизовать в middleware
	return true
}

func (r *Router) pick(u *UpdateCtx) (HandlerFunc, bool) {
	if cb := u.Update.CallbackQuery; cb != nil {
		data := cb.Data
		if h, ok := r.cbExact[data]; ok {
			return h, true
		}
		for p, f := range r.cbPrefix {
			if strings.HasPrefix(data, p) {
				vals := Parse(data)
				return func(uc *UpdateCtx) error { return f(uc, vals) }, true
			}
		}
		if r.def != nil {
			return r.def, true
		}
		return nil, false
	}
	if msg := u.Update.Message; msg != nil {
		if msg.IsCommand() {
			if h, ok := r.commands[msg.Command()]; ok {
				return h, true
			}
			if r.def != nil {
				return r.def, true
			}
			return nil, false
		}
		if h, ok := r.messages[msg.Text]; ok {
			return h, true
		}
		if r.def != nil {
			return r.def, true
		}
	}
	return nil, false
}
