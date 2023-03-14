package handlers

import (
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveFormAction(t *testing.T) {
	insertTestData(db)

	msg := buildMessage(TestUID3)
	reqenv := buildRequestEnv()
	fieldsFile := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: wizard.File{
			ID:       TestFileID,
			UniqueID: TestUniqueFileID,
		}},
	}

	saveFormAction(reqenv, msg, fieldsFile)

	checkRowsCount(t, 1, TestUID3, nil)
	itemFile := fetchItem(t, TestType)
	assert.Equal(t, TestFileID, *itemFile.FileID)
	assert.Equal(t, TestUniqueFileID, *itemFile.FileUniqueID)
	assert.Nil(t, itemFile.Text)

	fieldsText := fieldsFile
	objField := fieldsText.FindField(FieldObject)
	objField.Type = wizard.Text
	objField.Data = TestText

	saveFormAction(reqenv, msg, fieldsText)

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
	itemsRes := db.QueryRow("SELECT a.name, type, file_id, file_unique_id, text FROM items i JOIN aliases a ON i.alias = a.id WHERE uid = $1 AND type = $2", TestUID3, fieldType)
	var item queryResult
	err := itemsRes.Scan(&item.Name, &item.Type, &item.FileID, &item.FileUniqueID, &item.Text)
	assert.NoError(t, err)
	assert.Equal(t, TestAlias, item.Name)
	assert.Equal(t, fieldType, item.Type)
	return &item
}
