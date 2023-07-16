package handlers

import (
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHideAction(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldAlias,
			Data: test.Alias,
		},
	}

	aliasService := repo.NewAliasService(appenv)
	res, err := aliasService.List(test.UID)
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	handler := NewHideHandler(appenv, nil)
	handler.hideAction(reqenv, msg, fields)

	res, err = aliasService.List(test.UID)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}
