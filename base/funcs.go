package base

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
)

func GetCommandArgument(msg *tgbotapi.Message) string {
	return strings.TrimSpace(strings.TrimPrefix(msg.Text, "/"+msg.Command()))
}

func NewBotAPI(api *tgbotapi.BotAPI) *BotAPI {
	return &BotAPI{internal: api}
}

type MessageCustomizer func(msgConfig *tgbotapi.MessageConfig)

var (
	noOpCustomizer     MessageCustomizer = func(msgConfig *tgbotapi.MessageConfig) {}
	markdownCustomizer MessageCustomizer = func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
	}
)

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
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(buttons...),
	)
	keyboard.OneTimeKeyboard = true

	bot.ReplyWithMessageCustomizer(msg, text, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ReplyMarkup = keyboard
	})
}

func (bot *BotAPI) ReplyWithInlineKeyboard(msg *tgbotapi.Message, text string, buttons []InlineButton) {
	tgButtons := funk.Map(buttons, func(btn InlineButton) tgbotapi.InlineKeyboardButton {
		return tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.Data)
	}).([]tgbotapi.InlineKeyboardButton)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgButtons...),
	)

	bot.ReplyWithMessageCustomizer(msg, text, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ReplyMarkup = keyboard
	})
}

func (bot *BotAPI) Request(c tgbotapi.Chattable) error {
	_, err := bot.internal.Request(c)
	return err
}

func (reqenv *RequestEnv) Reply(text string) {
	reqenv.Bot.Reply(reqenv.Message, text)
}

func (reqenv *RequestEnv) ReplyWithMarkdown(text string) {
	reqenv.Bot.ReplyWithMarkdown(reqenv.Message, text)
}

func (reqenv *RequestEnv) ReplyWithKeyboard(text string, options []string) {
	reqenv.Bot.ReplyWithKeyboard(reqenv.Message, text, options)
}

func (reqenv *RequestEnv) ReplyWithInlineKeyboard(text string, buttons []InlineButton) {
	reqenv.Bot.ReplyWithInlineKeyboard(reqenv.Message, text, buttons)
}
