package help

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHelpCommand(t *testing.T) {
	appenv := &base.ApplicationEnv{
		Bot: &base.FakeBotAPI{},
		Ctx: context.Background(),
	}
	reqenv := test.BuildRequestEnv()
	msg := &tgbotapi.Message{
		MessageID: 0,
		From:      &tgbotapi.User{FirstName: "Test"},
		Chat:      &tgbotapi.Chat{Type: "private"},
	}

	handler := NewCommandHandler(appenv)
	handler.Handle(reqenv, msg)

	bot := appenv.Bot.(*base.FakeBotAPI)
	assert.Contains(t, bot.GetOutput().([]string)[0], "Hello, *Test*!")
}

func TestHelpCallback(t *testing.T) {
	appenv := &base.ApplicationEnv{
		Bot: &base.FakeBotAPI{},
		Ctx: context.Background(),
	}
	reqenv := test.BuildRequestEnv()
	query := &tgbotapi.CallbackQuery{
		Data:    callbackDataPrefix + string(groupsHelpKey),
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{}},
	}

	InitMessages(0, 0, "", []base.MessageHandler{NewCommandHandler(appenv)})
	handler := NewCallbackHandler(appenv)
	handler.Handle(reqenv, query)

	bot := appenv.Bot.(*base.FakeBotAPI)
	sentChattable := bot.GetOutput().([]tgbotapi.Chattable)[0].(tgbotapi.EditMessageTextConfig)
	assert.Contains(t, sentChattable.Text, "— `/help`")
}
