package handlers

import (
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchModeAction(t *testing.T) {
	insertTestData(db)

	_, opts := settings.FetchUserOptions(ctx, db, TestUID, "")
	assert.False(t, opts.SubstrSearchEnabled)

	reqenv := buildRequestEnv()
	msg := buildMessage(TestUID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldSubstrSearchEnabled,
			Data: Yes,
		},
	}
	searchModeAction(reqenv, msg, fields)

	_, opts = settings.FetchUserOptions(ctx, db, TestUID, "")
	assert.True(t, opts.SubstrSearchEnabled)
}
