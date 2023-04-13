package base

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
)

// GetCommandArgument extracts a command argument from the text of a message.
// For example:
// 	* "/foo bar" will result in "bar"
// 	* "/foo" will result in ""
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
	if bot.DummyMode {
		return "DummyMode"
	}
	return bot.internal.Self.UserName
}

// ReplyWithMessageCustomizer is the most common method to send text messages as a reply. Use this method if you want
// to change several options like a message in Markdown with an inline keyboard.
func (bot *BotAPI) ReplyWithMessageCustomizer(msg *tgbotapi.Message, text string, customizer MessageCustomizer) {
	if bot.DummyMode {
		return
	}
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

// Reply with just a text message, without any customizations.
func (bot *BotAPI) Reply(msg *tgbotapi.Message, text string) {
	bot.ReplyWithMessageCustomizer(msg, text, noOpCustomizer)
}

func (bot *BotAPI) ReplyWithMarkdown(msg *tgbotapi.Message, text string) {
	bot.ReplyWithMessageCustomizer(msg, text, markdownCustomizer)
}

// ReplyWithKeyboard uses a one time reply keyboard.
// https://core.telegram.org/bots/api#replykeyboardmarkup
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

// ReplyWithInlineKeyboard attaches an inline keyboard to the message.
// https://core.telegram.org/bots/api#inlinekeyboardmarkup
func (bot *BotAPI) ReplyWithInlineKeyboard(msg *tgbotapi.Message, text string, buttons []tgbotapi.InlineKeyboardButton) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)
	bot.ReplyWithMessageCustomizer(msg, text, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ReplyMarkup = keyboard
	})
}

// Request is the most common method that can be used to send any request to Telegram.
// This is a simple wrapper around [tgbotapi.BotAPI.Request].
func (bot *BotAPI) Request(c tgbotapi.Chattable) error {
	if bot.DummyMode {
		return nil
	}
	_, err := bot.internal.Request(c)
	return err
}
