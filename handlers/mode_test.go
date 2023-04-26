package handlers

import (
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchModeAction(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	userService := repo.NewUserService(appenv)

	_, opts := userService.FetchUserOptions(test.UID, "")
	assert.False(t, opts.(*dto.UserOptions).SubstrSearchEnabled)

	reqenv := test.BuildRequestEnv()
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldSubstrSearchEnabled,
			Data: Yes,
		},
	}
	handler := NewSearchModeHandler(appenv, nil)
	handler.searchModeAction(reqenv, msg, fields)

	_, opts = userService.FetchUserOptions(test.UID, "")
	assert.True(t, opts.(*dto.UserOptions).SubstrSearchEnabled)
}
