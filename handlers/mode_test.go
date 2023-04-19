package handlers

import (
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchModeAction(t *testing.T) {
	test.InsertTestData(db)

	userService := repo.NewUserService(ctx, db)

	_, opts := userService.FetchUserOptions(test.UID, "")
	assert.False(t, opts.SubstrSearchEnabled)

	reqenv := test.BuildRequestEnv(db)
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldSubstrSearchEnabled,
			Data: Yes,
		},
	}
	searchModeAction(reqenv, msg, fields)

	_, opts = userService.FetchUserOptions(test.UID, "")
	assert.True(t, opts.SubstrSearchEnabled)
}
