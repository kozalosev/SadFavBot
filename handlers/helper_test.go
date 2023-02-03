package handlers

import (
	"github.com/kozalosev/SadFavBot/wizard"
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
