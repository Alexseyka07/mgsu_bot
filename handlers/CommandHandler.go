package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler struct {
	botHandler BotHandler
}

func NewCommandHandler(botHandler *BotHandler) CommandHandler {
	return CommandHandler{
		botHandler: *botHandler,
	}
}

func (h *CommandHandler) CommandHandler(update *tgbotapi.Update) bool {
	if update.Message != nil && update.Message.IsCommand() {
		h.handleCommand(update.Message)
		return true
	}
	return false
}

func (h *CommandHandler) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		h.handleStartCommand(message)
	}

}

func (h *CommandHandler) handleStartCommand(message *tgbotapi.Message) {
	msg := "Привет! Нажми кнопку ниже и получи информацию о своем месте в конкурсном списке."
	buttons := []string{"Получить"}
	commands := h.botHandler.SetKeyboardButtons(buttons, 2)

	h.botHandler.SendTextMessageWithKeyboardMarkup(message.Chat.ID, msg, commands)
}
