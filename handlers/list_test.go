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

	reqenv := test.BuildRequestEnv(db)
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldFavsOrPackages,
			Data: Favs,
		},
	}

	listAction(reqenv, msg, fields)

	bot := reqenv.Bot.(*base.FakeBotAPI)
	sentMessage := bot.GetOutput().(string)
	lines := strings.Split(sentMessage, "\n")

	heading := lines[0]
	assert.Contains(t, heading, "favs")

	list := lines[2:]
	assert.Len(t, list, 2)
	assert.Contains(t, list[0], test.Alias)
	assert.Contains(t, list[1], test.Alias2)
}

func TestListActionPackages(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	reqenv := test.BuildRequestEnv(db)
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldFavsOrPackages,
			Data: Packages,
		},
	}

	listAction(reqenv, msg, fields)

	bot := reqenv.Bot.(*base.FakeBotAPI)
	sentMessage := bot.GetOutput().(string)
	lines := strings.Split(sentMessage, "\n")

	heading := lines[0]
	assert.Contains(t, heading, "packages")

	list := lines[2:]
	assert.Len(t, list, 1)
	assert.Contains(t, list[0], test.PackageFullName)
}
