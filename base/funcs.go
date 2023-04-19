package base

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
)

// GetCommandArgument extracts a command argument from the text of a message.
// For example:
//   - "/foo bar" will result in "bar"
//   - "/foo" will result in ""
func GetCommandArgument(msg *tgbotapi.Message) string {
	return strings.TrimSpace(strings.TrimPrefix(msg.Text, "/"+msg.Command()))
}

func NewBotAPI(api *tgbotapi.BotAPI) *BotAPI {
	return &BotAPI{internal: api}
}

// MessageCustomizer is a function that can change the message before it will be sent to Telegram.
// See [BotAPI.ReplyWithMessageCustomizer] for more information.
type MessageCustomizer func(msgConfig *tgbotapi.MessageConfig)

var (
	noOpCustomizer     MessageCustomizer = func(msgConfig *tgbotapi.MessageConfig) {}
	markdownCustomizer MessageCustomizer = func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
	}
)

func (bot *BotAPI) GetName() string {
	return bot.internal.Self.UserName
}

func (bot *BotAPI) ReplyWithMessageCustomizer(msg *tgbotapi.Message, text string, customizer MessageCustomizer) {
	if len(text) == 0 {
		log.Error("Empty reply for the message: " + msg.Text)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID
	customizer(&reply)
	if _, err := bot.internal.Send(reply); err != nil {
		log.Errorln(err)
	}
}

func (bot *BotAPI) Reply(msg *tgbotapi.Message, text string) {
	bot.ReplyWithMessageCustomizer(msg, text, noOpCustomizer)
}

func (bot *BotAPI) ReplyWithMarkdown(msg *tgbotapi.Message, text string) {
	bot.ReplyWithMessageCustomizer(msg, text, markdownCustomizer)
}

func (bot *BotAPI) ReplyWithKeyboard(msg *tgbotapi.Message, text string, options []string) {
	buttons := funk.Map(options, func(s string) tgbotapi.KeyboardButton {
		return tgbotapi.NewKeyboardButton(s)
	}).([]tgbotapi.KeyboardButton)
	keyboard := tgbotapi.NewOneTimeReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(buttons...),
	)

	bot.ReplyWithMessageCustomizer(msg, text, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ReplyMarkup = keyboard
	})
}

func (bot *BotAPI) ReplyWithInlineKeyboard(msg *tgbotapi.Message, text string, buttons []tgbotapi.InlineKeyboardButton) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)
	bot.ReplyWithMessageCustomizer(msg, text, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ReplyMarkup = keyboard
	})
}

// Request is a simple wrapper around [tgbotapi.BotAPI.Request].
func (bot *BotAPI) Request(c tgbotapi.Chattable) error {
	_, err := bot.internal.Request(c)
	return err
}
