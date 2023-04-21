package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractItemValues_File(t *testing.T) {
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: test.Alias},
		&wizard.Field{Name: FieldObject, Type: test.Type, Data: wizard.File{
			ID:       test.FileID,
			UniqueID: test.UniqueFileID,
		}},
	}

	alias, fav := extractFavInfo(fields)

	assert.Equal(t, test.Alias, alias)
	assert.Equal(t, test.Type, fav.Type)
	assert.Equal(t, test.FileID, fav.File.ID)
	assert.Equal(t, test.UniqueFileID, fav.File.UniqueID)
	assert.Empty(t, fav.Text)
}

func TestExtractItemValues_Text(t *testing.T) {
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: test.Alias},
		&wizard.Field{Name: FieldObject, Type: wizard.Text, Data: test.Text},
	}

	alias, fav := extractFavInfo(fields)

	assert.Equal(t, test.Alias, alias)
	assert.Equal(t, wizard.Text, fav.Type)
	assert.Equal(t, test.Text, *fav.Text)
	assert.Empty(t, fav.File.ID)
	assert.Empty(t, fav.File.UniqueID)
}

func TestExtractItemValues_MismatchError(t *testing.T) {
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: test.Alias},
		&wizard.Field{Name: FieldObject, Type: test.Type, Data: test.Text},
	}

	alias, fav := extractFavInfo(fields)

	assert.Empty(t, alias)
	assert.Nil(t, fav)
}

func buildMessage(uid int64) *tgbotapi.Message {
	return &tgbotapi.Message{
		From: &tgbotapi.User{ID: uid},
	}
}
