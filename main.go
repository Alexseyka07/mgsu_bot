package main

import (
	"bot/config"
	"bot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(config.TG_TOKEN)
	if err != nil {
		panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 50

	updates := bot.GetUpdatesChan(u)

	bot_handler := handlers.NewBotHandler(&updates, bot)

	command_handler := handlers.NewCommandHandler(&bot_handler)
	mgsu_handler := handlers.NewMgsuHandler(&bot_handler)

	bot_handler.AddHandler(command_handler.CommandHandler)
	bot_handler.AddHandler(mgsu_handler.MgsuHandler)

	// Запускаем мониторинг МГСУ
	mgsu_handler.StartMonitoring()

	println("Bot is running")

	bot_handler.MessagesHandler()

}
