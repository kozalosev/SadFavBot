package common

import (
	log "log/slog"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/selfdestruct"
	"github.com/loctools/go-l10n/loc"
)

// ReplyOption changes how the self-destroying repliers behave.
type ReplyOption func(*replyOptions)

type replyOptions struct {
	toOriginal bool
	delay      time.Duration
}

// ToOriginal makes the replier answer to the message the command was sent in reply to (see [ReplyToOriginal]).
// The command itself is still the message deleted together with the reply.
func ToOriginal(o *replyOptions) { o.toOriginal = true }

// PossiblySelfDestroyingReplier replies as is in private chats and to ephemeral messages, which disappear by themselves,
// and sends a self-destroying message to a group where the bot may delete messages of other users: after the delay,
// the reply is replaced with the warning, and then it's deleted together with the command. Where the bot may not,
// nothing is deleted, so that the command isn't left alone.
func PossiblySelfDestroyingReplier(appenv *base.ApplicationEnv, reqenv *base.RequestEnv, msg *tgbotapi.Message, opts ...ReplyOption) func(string) {
	return func(statusKey string) {
		ReplyPossiblySelfDestroying(appenv, reqenv.Lang, msg, reqenv.Lang.Tr(statusKey), base.NoOpCustomizer, opts...)
	}
}

// ReplyPossiblySelfDestroying is the same as PossiblySelfDestroyingReplier, but for an arbitrary text. lang is used
// for the warning.
func ReplyPossiblySelfDestroying(appenv *base.ApplicationEnv, lang *loc.Context, msg *tgbotapi.Message, text string, customizer base.MessageCustomizer, opts ...ReplyOption) {
	options := replyOptions{delay: selfDeletionDelay}
	for _, opt := range opts {
		opt(&options)
	}
	replyTo := msg
	if options.toOriginal {
		replyTo = ReplyToOriginal(msg)
	}

	// without the rights, the command would be left alone
	if msg.Chat.IsPrivate() || msg.EphemeralMessageID != 0 || !base.CanDeleteMessagesOfOthers(appenv.Bot, msg.Chat.ID) {
		appenv.Bot.ReplyWithMessageCustomizer(replyTo, text, customizer)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	customizer(&reply)
	reply.ReplyParameters.MessageID = replyTo.MessageID
	sent, err := appenv.Bot.Send(reply)
	if err != nil {
		log.Error("Failed to send a self-destroying reply",
			logconst.FieldFunc, "ReplyPossiblySelfDestroying",
			logconst.FieldCalledObject, "Bot",
			logconst.FieldCalledMethod, "Send",
			logconst.FieldError, err)
		return
	}
	selfdestruct.Schedule(appenv.Bot, selfdestruct.Task{
		Kind:           selfdestruct.KindDelete,
		ChatID:         msg.Chat.ID,
		MessageID:      sent.MessageID,
		WithMessageIDs: []int{msg.MessageID},
		Warning:        deletionWarning(lang, sent.MessageID),
	}, options.delay)
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
