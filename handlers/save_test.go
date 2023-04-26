package handlers

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveFormAction(t *testing.T) {
	test.InsertTestData(db)

	msg := buildMessage(test.UID3)
	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	fieldsFile := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: test.Alias},
		&wizard.Field{Name: FieldObject, Type: test.Type, Data: wizard.File{
			ID:       test.FileID,
			UniqueID: test.UniqueFileID,
		}},
	}

	handler := NewSaveHandler(appenv, nil)
	handler.saveFormAction(reqenv, msg, fieldsFile)

	test.CheckRowsCount(t, db, 1, test.UID3, nil)
	itemFile := fetchItem(t, test.Type)
	assert.Equal(t, test.FileID, *itemFile.FileID)
	assert.Equal(t, test.UniqueFileID, *itemFile.FileUniqueID)
	assert.Nil(t, itemFile.Text)

	fieldsText := fieldsFile
	objField := fieldsText.FindField(FieldObject)
	objField.Type = wizard.Text
	objField.Data = test.Text

	handler.saveFormAction(reqenv, msg, fieldsText)

	test.CheckRowsCount(t, db, 2, test.UID3, nil)
	itemText := fetchItem(t, wizard.Text)
	assert.Nil(t, itemText.FileID)
	assert.Nil(t, itemText.FileUniqueID)
	assert.Equal(t, test.Text, *itemText.Text)
}

type queryResult struct {
	Name         string
	Type         wizard.FieldType
	FileID       *string
	FileUniqueID *string
	Text         *string
}

func fetchItem(t *testing.T, fieldType wizard.FieldType) *queryResult {
	itemsRes := db.QueryRow(ctx, "SELECT a.name, type, file_id, file_unique_id, t.text FROM favs f JOIN aliases a ON f.alias_id = a.id LEFT JOIN texts t on t.id = f.text_id WHERE uid = $1 AND type = $2", test.UID3, fieldType)
	var item queryResult
	err := itemsRes.Scan(&item.Name, &item.Type, &item.FileID, &item.FileUniqueID, &item.Text)
	assert.NoError(t, err)
	assert.Equal(t, test.Alias, item.Name)
	assert.Equal(t, fieldType, item.Type)
	return &item
}
