package common

import (
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/selfdestruct"
)

// ReplyOption changes how the self-destroying repliers behave.
type ReplyOption func(*replyOptions)

type replyOptions struct {
	toOriginal bool
}

// ToOriginal makes the replier answer to the message the command was sent in reply to (see [ReplyToOriginal]).
// The command itself is still the message deleted together with the reply.
func ToOriginal(o *replyOptions) { o.toOriginal = true }

// PossiblySelfDestroyingReplier replies as is in private chats and to ephemeral messages, which disappear by themselves,
// and sends a self-destroying message to a group where the bot may delete messages of other users, deleting
// the command together with the reply. Where it may not, nothing is deleted, so that the command isn't left alone.
func PossiblySelfDestroyingReplier(appenv *base.ApplicationEnv, reqenv *base.RequestEnv, msg *tgbotapi.Message, opts ...ReplyOption) func(string) {
	return func(statusKey string) {
		ReplyPossiblySelfDestroying(appenv, msg, reqenv.Lang.Tr(statusKey), base.NoOpCustomizer, opts...)
	}
}

// ReplyPossiblySelfDestroying is the same as PossiblySelfDestroyingReplier, but for an arbitrary text.
func ReplyPossiblySelfDestroying(appenv *base.ApplicationEnv, msg *tgbotapi.Message, text string, customizer base.MessageCustomizer, opts ...ReplyOption) {
	var options replyOptions
	for _, opt := range opts {
		opt(&options)
	}
	replyTo := msg
	if options.toOriginal {
		replyTo = ReplyToOriginal(msg)
	}

	// without the rights, the command would be left alone
	if msg.Chat.IsPrivate() || msg.EphemeralMessageID != 0 || !canDeleteMessagesOfOthers(appenv.Bot, msg.Chat.ID) {
		appenv.Bot.ReplyWithMessageCustomizer(replyTo, text, customizer)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	customizer(&reply)
	replyDestroyingImpl(appenv.Bot, replyTo, reply)
	selfdestruct.DeleteAfter(appenv.Bot, msg.Chat.ID, msg.MessageID, selfDeletionDelay)
}

func canDeleteMessagesOfOthers(bot base.ExtendedBotAPI, chatID int64) bool {
	checker, ok := bot.(base.DeletionRightsChecker)
	return ok && checker.CanDeleteMessages(chatID)
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
