package handlers

import (
	"strings"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"

	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
)

func TestListActionFavs(t *testing.T) {
	test.InsertTestData(db)

	testListActionImpl(t, Favs, func(list []string) {
		assert.Len(t, list, 2)
		assert.Contains(t, list[0], test.Alias)
		assert.Contains(t, list[1], test.Alias2)
	})
}

func TestListActionPackages(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	testListActionImpl(t, Packages, func(list []string) {
		assert.Len(t, list, 1)
		assert.Contains(t, list[0], test.PackageFullName)
	})
}

func testListActionImpl(t *testing.T, favsOrPackages string, assertionsFunc func(list []string)) {
	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		test.NewTextField(FieldFavsOrPackages, favsOrPackages),
		test.NewTextField(FieldGrep, ""),
	}

	handler := NewListHandler(appenv, nil)
	handler.listAction(reqenv, msg, fields)

	bot := appenv.Bot.(*base.FakeBotAPI)
	sentMessages := bot.GetOutput().([]string)
	assert.Len(t, sentMessages, 1)
	sentMessage := sentMessages[0]
	lines := strings.Split(sentMessage, "\n")

	heading := lines[0]
	assert.Contains(t, heading, strings.ToLower(favsOrPackages))

	list := lines[2:]
	assertionsFunc(list)
}

func TestBuildEditRequest(t *testing.T) {
	chat := tgbotapi.Chat{ID: 100}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➡️", "data")))

	req := buildEditRequest(&tgbotapi.Message{MessageID: 5, Chat: chat}, 7, "text", &keyboard)
	edit, ok := req.(tgbotapi.EditMessageTextConfig)
	assert.True(t, ok)
	assert.Equal(t, 5, edit.MessageID)
	assert.NotNil(t, edit.ReplyMarkup)

	req = buildEditRequest(&tgbotapi.Message{MessageID: 5, Chat: chat}, 7, "text", nil)
	edit, ok = req.(tgbotapi.EditMessageTextConfig)
	assert.True(t, ok)
	assert.Nil(t, edit.ReplyMarkup)

	req = buildEditRequest(&tgbotapi.Message{Chat: chat, EphemeralMessageID: 9}, 7, "text", &keyboard)
	ephemeral, ok := req.(tgbotapi.EditEphemeralMessageTextConfig)
	assert.True(t, ok)
	assert.Equal(t, 9, ephemeral.EphemeralMessageID)
	assert.Equal(t, int64(7), ephemeral.ReceiverUserID)
	assert.Equal(t, &keyboard, ephemeral.ReplyMarkup)
}
