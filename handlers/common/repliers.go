package common

import (
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/selfdestruct"
)

// PossiblySelfDestroyingReplier replies as is in private chats and to ephemeral messages, which disappear by themselves,
// and sends a self-destroying message to a group.
func PossiblySelfDestroyingReplier(appenv *base.ApplicationEnv, reqenv *base.RequestEnv, msg *tgbotapi.Message) func(string) {
	if msg.Chat.IsPrivate() || msg.EphemeralMessageID != 0 {
		return base.NewReplier(appenv, reqenv, msg)
	}

	return func(statusKey string) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, reqenv.Lang.Tr(statusKey))
		replyDestroyingImpl(appenv.Bot, msg, reply)
	}
}

// ReplyPossiblySelfDestroying is the same as PossiblySelfDestroyingReplier, but for an arbitrary text.
func ReplyPossiblySelfDestroying(appenv *base.ApplicationEnv, msg *tgbotapi.Message, text string, customizer base.MessageCustomizer) {
	if msg.Chat.IsPrivate() || msg.EphemeralMessageID != 0 {
		appenv.Bot.ReplyWithMessageCustomizer(msg, text, customizer)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	customizer(&reply)
	replyDestroyingImpl(appenv.Bot, msg, reply)
}

// SelfDeletionDelay is how long a self-destroying message lives.
func SelfDeletionDelay() time.Duration {
	return selfDeletionDelay
}

// ReplyToOriginal returns the message to pass to the repliers to answer to the message the command was sent in reply to
// instead of the command itself. The message is returned as is if it's not a reply or if it's ephemeral, because
// an ephemeral message can be answered only to the command that triggered it.
func ReplyToOriginal(msg *tgbotapi.Message) *tgbotapi.Message {
	if msg.ReplyToMessage == nil || msg.EphemeralMessageID != 0 {
		return msg
	}
	original := *msg
	original.MessageID = msg.ReplyToMessage.MessageID
	return &original
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
	selfdestruct.SendAndDelete(bot, msgToSend, selfDeletionDelay)
}
