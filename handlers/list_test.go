package handlers

import (
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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
		&wizard.Field{
			Name: FieldFavsOrPackages,
			Data: favsOrPackages,
		},
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
