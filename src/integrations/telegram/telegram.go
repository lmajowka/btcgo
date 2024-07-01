package telegram

import (
    "log"
    "os"
    "github.com/go-telegram-bot-api/telegram-bot-api"
)

func TelegramMessage(mensagem string) {

    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN não está definida")
    }

    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panic(err)
    }

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)
    if err != nil {
        log.Panic(err)
    }
    for update := range updates {
        if update.Message != nil {
            log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

            msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Olá, você disse: "+update.Message.Text)
            msg.ReplyToMessageID = update.Message.MessageID

            bot.Send(msg)
        }
    }
}
