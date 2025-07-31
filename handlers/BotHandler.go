package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot      tgbotapi.BotAPI
	updates  tgbotapi.UpdatesChannel
	handlers []func(*tgbotapi.Update) bool
}

func NewBotHandler(updates *tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) BotHandler {
	return BotHandler{
		bot:      *bot,
		updates:  *updates,
		handlers: []func(*tgbotapi.Update) bool{},
	}
}

func (b *BotHandler) AddHandler(handler func(*tgbotapi.Update) bool) {
	b.handlers = append(b.handlers, handler)
}

func (b *BotHandler) MessagesHandler() {
	for update := range b.updates {
		if update.Message != nil {
			for _, handler := range b.handlers {
				if handler(&update) {
					break
				}
			}
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		}
		if update.CallbackQuery != nil {
			for _, handler := range b.handlers {
				if handler(&update) {
					break
				}
			}
			log.Printf("[%s] %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
		}
	}
}

func (b *BotHandler) SendTextMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (b *BotHandler) SendTextMessageWithMarkup(chatID int64, text string, replyMarkup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = replyMarkup
	b.bot.Send(msg)
}

func (b *BotHandler) SendTextMessageWithKeyboardMarkup(chatID int64, text string, replyMarkup tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = replyMarkup
	b.bot.Send(msg)
}

func (b *BotHandler) SetKeyboardButtons(buttons []string, coloumsCount int) tgbotapi.ReplyKeyboardMarkup {
	var keyboard [][]tgbotapi.KeyboardButton
	var row []tgbotapi.KeyboardButton
	for i, button := range buttons {
		row = append(row, tgbotapi.NewKeyboardButton(button))
		if (i+1)%coloumsCount == 0 {
			keyboard = append(keyboard, row)
			row = []tgbotapi.KeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	return tgbotapi.NewReplyKeyboard(keyboard...)
}

func (b *BotHandler) SetInlineKeybordButtons(buttons map[string]string, columnsCount int) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	i := 0
	for title, data := range buttons {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(title, data))
		i++
		if i%columnsCount == 0 {
			keyboard = append(keyboard, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

func (b *BotHandler) SendTextMessageWithImage(chatID int64, text string, imagePath string) {
	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	msg.Caption = text
	if _, err := b.bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
