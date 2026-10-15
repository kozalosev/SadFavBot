package common

import (
	log "log/slog"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
)

func PossiblySelfDestroyingReplier(appenv *base.ApplicationEnv, reqenv *base.RequestEnv, msg *tgbotapi.Message) func(string) {
	if msg.Chat.IsPrivate() {
		return base.NewReplier(appenv, reqenv, msg)
	}

	return func(statusKey string) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, reqenv.Lang.Tr(statusKey))
		replyDestroyingImpl(appenv.Bot, msg, reply)
	}
}

func ReplyPossiblySelfDestroying(appenv *base.ApplicationEnv, msg *tgbotapi.Message, text string, customizer base.MessageCustomizer) {
	if msg.Chat.IsPrivate() {
		appenv.Bot.ReplyWithMessageCustomizer(msg, text, customizer)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	customizer(&reply)
	replyDestroyingImpl(appenv.Bot, msg, reply)
}

func SingleRowInlineKeyboardCustomizer(buttons []tgbotapi.InlineKeyboardButton) func(*tgbotapi.MessageConfig) {
	if len(buttons) > 0 {
		return func(msgConfig *tgbotapi.MessageConfig) {
			msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(buttons...),
			)
		}
	}
	return base.NoOpCustomizer
}

func replyDestroyingImpl(bot base.ExtendedBotAPI, srcMsg *tgbotapi.Message, msgToSend tgbotapi.MessageConfig) {
	msgToSend.ReplyParameters.MessageID = srcMsg.MessageID

	if sentMsg, err := bot.Send(msgToSend); err == nil {
		go func() {
			time.Sleep(selfDeletionDelay)
			if err := bot.Request(tgbotapi.NewDeleteMessage(sentMsg.Chat.ID, sentMsg.MessageID)); err != nil {
				log.Error("Couldn't delete a self-destroying message",
					logconst.FieldFunc, "replyDestroyingImpl",
					logconst.FieldError, err)
			}
		}()
	} else {
		log.Error("Failed to call ExtendedBotAPI.Send",
			logconst.FieldFunc, "replyDestroyingImpl",
			logconst.FieldCalledObject, "ExtendedBotAPI",
			logconst.FieldCalledMethod, "Send",
			logconst.FieldError, err)
	}
}
