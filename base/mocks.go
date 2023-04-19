package base

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type callType byte

const (
	message callType = iota
	request
)

// FakeBotAPI is a mock for the [BotAPI] struct.
// Use the GetOutput() method to get either the text of the sent message, or the request itself.
type FakeBotAPI struct {
	sentMessage string
	sentRequest tgbotapi.Chattable
	callType    callType
}

func (bot *FakeBotAPI) GetName() string {
	return "TestMockBotAPI"
}

func (bot *FakeBotAPI) ReplyWithMessageCustomizer(_ *tgbotapi.Message, text string, _ MessageCustomizer) {
	bot.reply(text)
}
func (bot *FakeBotAPI) Reply(_ *tgbotapi.Message, text string)             { bot.reply(text) }
func (bot *FakeBotAPI) ReplyWithMarkdown(_ *tgbotapi.Message, text string) { bot.reply(text) }
func (bot *FakeBotAPI) ReplyWithKeyboard(_ *tgbotapi.Message, text string, _ []string) {
	bot.reply(text)
}
func (bot *FakeBotAPI) ReplyWithInlineKeyboard(_ *tgbotapi.Message, text string, _ []tgbotapi.InlineKeyboardButton) {
	bot.reply(text)
}

func (bot *FakeBotAPI) reply(text string) {
	bot.callType = message
	bot.sentMessage = text
}

func (bot *FakeBotAPI) Request(c tgbotapi.Chattable) error {
	bot.callType = request
	bot.sentRequest = c
	return nil
}

// GetOutput returns either a string after usage of Reply*() methods or a [tgbotapi.Chattable] after Request()
func (bot *FakeBotAPI) GetOutput() interface{} {
	switch bot.callType {
	case message:
		return bot.sentMessage
	case request:
		return bot.sentRequest
	default:
		return nil
	}
}
