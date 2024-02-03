package common

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractItemValues_File(t *testing.T) {
	fields := wizard.Fields{
		test.NewTextField(FieldAlias, test.Alias),
		&wizard.Field{Name: FieldObject, Type: test.Type, Data: wizard.File{
			ID:       test.FileID,
			UniqueID: test.UniqueFileID,
		}},
	}

	alias, fav := ExtractFavInfo(fields)

	assert.Equal(t, test.Alias, alias)
	assert.Equal(t, test.Type, fav.Type)
	assert.Equal(t, test.FileID, fav.File.ID)
	assert.Equal(t, test.UniqueFileID, fav.File.UniqueID)
	assert.Empty(t, fav.Text)
	assert.Empty(t, fav.Location)
}

func TestExtractItemValues_Text(t *testing.T) {
	fields := wizard.Fields{
		test.NewTextField(FieldAlias, test.Alias),
		test.NewTextField(FieldObject, test.Text),
	}

	alias, fav := ExtractFavInfo(fields)

	assert.Equal(t, test.Alias, alias)
	assert.Equal(t, wizard.Text, fav.Type)
	assert.Equal(t, test.Text, fav.Text.Value)
	assert.Empty(t, fav.File.ID)
	assert.Empty(t, fav.File.UniqueID)
	assert.Empty(t, fav.Location)
}

func TestExtractItemValues_Location(t *testing.T) {
	loc := wizard.LocData{
		Latitude:  test.Latitude,
		Longitude: test.Longitude,
	}
	fields := wizard.Fields{
		test.NewTextField(FieldAlias, test.Alias),
		&wizard.Field{Name: FieldObject, Type: wizard.Location, Data: loc},
	}

	alias, fav := ExtractFavInfo(fields)

	assert.Equal(t, test.Alias, alias)
	assert.Equal(t, wizard.Location, fav.Type)
	assert.Equal(t, loc, *fav.Location)
	assert.Empty(t, fav.File.ID)
	assert.Empty(t, fav.File.UniqueID)
	assert.Empty(t, fav.Text)
}

func TestExtractItemValues_MismatchError(t *testing.T) {
	fields := wizard.Fields{
		test.NewTextField(FieldAlias, test.Alias),
		&wizard.Field{Name: FieldObject, Type: test.Type, Data: test.Text},
	}

	alias, fav := ExtractFavInfo(fields)

	assert.Empty(t, alias)
	assert.Nil(t, fav)
}
