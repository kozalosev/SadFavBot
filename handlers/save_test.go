package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveFormAction(t *testing.T) {
	insertTestData(db)

	reqenv := &base.RequestEnv{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: TestUID3},
		},
		Database: db,
		Bot:      &base.BotAPI{DummyMode: true},
		Lang:     loc.NewPool("en").GetContext("en"),
	}
	fieldsFile := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: wizard.File{
			ID:       TestFileID,
			UniqueID: TestUniqueFileID,
		}},
	}

	saveFormAction(reqenv, fieldsFile)

	checkRowsCount(t, 1, TestUID3, nil)
	itemFile := fetchItem(t, TestType)
	assert.Equal(t, TestFileID, *itemFile.FileID)
	assert.Equal(t, TestUniqueFileID, *itemFile.FileUniqueID)
	assert.Nil(t, itemFile.Text)

	fieldsText := fieldsFile
	objField := fieldsText.FindField(FieldObject)
	objField.Type = wizard.Text
	objField.Data = TestText

	saveFormAction(reqenv, fieldsText)

	checkRowsCount(t, 2, TestUID3, nil)
	itemText := fetchItem(t, wizard.Text)
	assert.Nil(t, itemText.FileID)
	assert.Nil(t, itemText.FileUniqueID)
	assert.Equal(t, TestText, *itemText.Text)
}

type queryResult struct {
	Name         string
	Type         wizard.FieldType
	FileID       *string
	FileUniqueID *string
	Text         *string
}

func fetchItem(t *testing.T, fieldType wizard.FieldType) *queryResult {
	itemsRes := db.QueryRow("SELECT alias, type, file_id, file_unique_id, text FROM items WHERE uid = $1 AND type = $2", TestUID3, fieldType)
	var item queryResult
	err := itemsRes.Scan(&item.Name, &item.Type, &item.FileID, &item.FileUniqueID, &item.Text)
	assert.NoError(t, err)
	assert.Equal(t, TestAlias, item.Name)
	assert.Equal(t, fieldType, item.Type)
	return &item
}
