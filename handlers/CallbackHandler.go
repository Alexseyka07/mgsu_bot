package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type CallbackHandler struct {
	botHandler BotHandler
}

func NewCallbackHandler(botHandler *BotHandler) CallbackHandler {
	return CallbackHandler{
		botHandler: *botHandler,
	}
}
func (h *CallbackHandler) CallbackHandler(update *tgbotapi.Update) bool {
	if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
		return true
	}
	return false
}

func (h *CallbackHandler) handleCallback(callback *tgbotapi.CallbackQuery) {
	switch callback.Data {
	case "wireguard":
		h.handleWireGuard(callback)
	case "vless":
		h.botHandler.SendTextMessage(callback.Message.Chat.ID, "Vless - это современный протокол, который обеспечивает высокую скорость и безопасность. Он использует криптографические методы для защиты данных и является простым в настройке. Vless подходит для большинства устройств и операционных систем.")
	case "difference":
		h.botHandler.SendTextMessage(callback.Message.Chat.ID, "WireGuard и Vless - это два разных VPN-протокола. WireGuard - это современный протокол, который обеспечивает высокую скорость и безопасность, используя криптографические методы. Vless - это более новый протокол, который также обеспечивает высокую скорость и безопасность, но имеет некоторые отличия в архитектуре и")
	default:
		h.botHandler.SendTextMessage(callback.Message.Chat.ID, "Неизвестная команда")
	}
}
func (h *CallbackHandler) handleWireGuard(callback *tgbotapi.CallbackQuery) {
	msg := `WireGuard - это современный VPN-протокол, который обеспечивает высокую скорость и безопасность. Он использует криптографические методы для защиты данных и является простым в настройке. WireGuard подходит для большинства устройств и операционных систем.`

	h.botHandler.SendTextMessage(callback.Message.Chat.ID, msg)
}
