package handlers

import (
	"walki/internal/constants"
	"walki/internal/keyboards"
)

func (h *Handler) registerButtons() {
	h.RegisterButton(keyboards.ButtonTexts[constants.BtnRoutes], h.handleRoutes)
	h.RegisterButton(keyboards.ButtonTexts[constants.BtnProfile], h.handleProfile)
	h.RegisterButton(keyboards.ButtonTexts[constants.BtnHelp], h.handleHelp)
}
