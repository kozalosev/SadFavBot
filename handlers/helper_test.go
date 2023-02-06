package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractItemValues_File(t *testing.T) {
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: wizard.File{
			ID:       TestFileID,
			UniqueID: TestUniqueFileID,
		}},
	}

	res, ok := extractItemValues(fields)

	assert.True(t, ok)
	assert.Equal(t, TestAlias, res.Alias)
	assert.Equal(t, TestType, res.Type)
	assert.Equal(t, TestFileID, res.File.ID)
	assert.Equal(t, TestUniqueFileID, res.File.UniqueID)
	assert.Empty(t, res.Text)
}

func TestExtractItemValues_Text(t *testing.T) {
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldObject, Type: wizard.Text, Data: TestText},
	}

	res, ok := extractItemValues(fields)

	assert.True(t, ok)
	assert.Equal(t, TestAlias, res.Alias)
	assert.Equal(t, wizard.Text, res.Type)
	assert.Equal(t, TestText, res.Text)
	assert.Empty(t, res.File.ID)
	assert.Empty(t, res.File.UniqueID)
}

func TestExtractItemValues_MismatchError(t *testing.T) {
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: TestText},
	}

	res, ok := extractItemValues(fields)

	assert.False(t, ok)
	assert.Nil(t, res)
}

func buildRequestEnvWithMessage(uid int64) *base.RequestEnv {
	return &base.RequestEnv{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: uid},
		},
		Database: db,
		Bot:      &base.BotAPI{DummyMode: true},
		Lang:     loc.NewPool("en").GetContext("en"),
	}
}
