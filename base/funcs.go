package base

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

func (bot *BotAPI) Reply(msg *tgbotapi.Message, text string) {
	if bot.DummyMode {
		return
	}

	if len(text) == 0 {
		log.Error("Empty reply for a message: " + msg.Text)
		return
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID
	if _, err := bot.Send(reply); err != nil {
		panic(err)
	}
}

func (reqenv *RequestEnv) Reply(text string) {
	reqenv.Bot.Reply(reqenv.Message, text)
}
