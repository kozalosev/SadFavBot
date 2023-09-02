package handlers

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestRefHandler_refAction(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldObject,
			Type: test.Type,
			Data: wizard.File{UniqueID: test.UniqueFileID},
		},
	}

	handler := NewRefHandler(appenv, nil)
	handler.refAction(reqenv, msg, fields)

	bot := appenv.Bot.(*base.FakeBotAPI)
	sentMessages := bot.GetOutput().([]string)
	assert.Len(t, sentMessages, 1)
	sentMessage := sentMessages[0]
	lines := strings.Split(sentMessage, "\n")
	list := lines[2:]

	assert.Len(t, list, 2)
	assert.Contains(t, list[0], test.Alias)
	assert.Contains(t, list[1], test.Alias2)
}
