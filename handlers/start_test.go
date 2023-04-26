package handlers

import (
	"encoding/base64"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStart(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	bot := appenv.Bot.(*base.FakeBotAPI)
	stateStorage := &wizard.FakeStorage{}
	handler := NewStartHandler(appenv, stateStorage, StartEmbeddedHandlers{
		Language:       NewLanguageHandler(appenv, stateStorage),
		InstallPackage: NewInstallPackageHandler(appenv, stateStorage),
	})
	wizard.PopulateWizardDescriptors([]base.MessageHandler{
		handler,
		handler.embeddedHandlers.Language,
		handler.embeddedHandlers.InstallPackage,
	})

	var uid int64 = test.UID - 1
	msg := buildMessage(uid)

	assert.False(t, handler.CanHandle(reqenv, msg))
	msg.Text = "/start"
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Length: len(msg.Text)}}
	assert.True(t, handler.CanHandle(reqenv, msg))

	assert.False(t, userExists(uid))
	handler.Handle(reqenv, msg)
	assert.True(t, userExists(uid))

	sentMessages := bot.GetOutput().([]string)
	assert.Len(t, sentMessages, 1)
	assert.Contains(t, sentMessages[0], "language")

	bot.ClearOutput()
	handler.Handle(reqenv, msg)

	sentMessages = bot.GetOutput().([]string)
	assert.Len(t, sentMessages, 1)
	assert.Contains(t, sentMessages[0], "Hello")

	encodedPkgName := []byte(test.PackageFullName)
	msg.Text = "/start " + base64.StdEncoding.EncodeToString(encodedPkgName)
	bot.ClearOutput()
	handler.Handle(reqenv, msg)

	sentMessages = bot.GetOutput().([]string)
	assert.Len(t, sentMessages, 1)
	assert.Contains(t, sentMessages[0], "confirmation")
}

func userExists(uid int64) bool {
	var x int
	if err := db.QueryRow(ctx, "SELECT 1 FROM users WHERE uid = $1", uid).Scan(&x); err != nil {
		return false
	}
	return x == 1
}
