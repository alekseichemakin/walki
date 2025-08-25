package handlers

func (h *Handler) registerCommands() {
	h.RegisterCommand("start", h.handleStart)
	// Здесь можно зарегистрировать другие команды
}
