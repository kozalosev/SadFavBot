package repo

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testAlias = "test_alias"

func TestTrimCountSuffix(t *testing.T) {
	assert.Equal(t, testAlias, trimCountSuffix(testAlias+" (1)"))
	assert.Equal(t, testAlias, trimCountSuffix(testAlias))
	assert.Equal(t, testAlias+" (test)", trimCountSuffix(testAlias+" (test)"))
}

func TestTrimLinkSuffix(t *testing.T) {
	assert.Equal(t, testAlias, trimLinkSuffix(testAlias+" → "+testAlias))
	assert.Equal(t, testAlias, trimLinkSuffix(testAlias))
}

func TestTrimSuffix(t *testing.T) {
	assert.Equal(t, testAlias, trimSuffix(testAlias+" (1)"))
	assert.Equal(t, testAlias, trimSuffix(testAlias+" → "+testAlias))
	assert.Equal(t, testAlias, trimSuffix(testAlias))
}

func TestAliasService_ListWithCounts(t *testing.T) {
	test.InsertTestData(db)

	aliasService := NewAliasService(test.BuildApplicationEnv(db))
	aliases, err := aliasService.ListWithCounts(test.UID, "", "")

	assert.NoError(t, err)
	assert.Len(t, aliases.Items, 2)
	assert.False(t, aliases.HasNextPage)
	assert.Contains(t, aliases.Items, test.Alias+" (2)")
	assert.Contains(t, aliases.Items, test.Alias2+" (1)")
}

func TestAliasService_List_noRows(t *testing.T) {
	test.InsertTestData(db)

	aliasService := NewAliasService(test.BuildApplicationEnv(db))
	aliases, err := aliasService.List(test.UID - 1)

	assert.NoError(t, err)
	assert.Len(t, aliases, 0)
}

func TestAliasService_ListWithCounts_noHidden(t *testing.T) {
	test.InsertTestData(db)
	appEnv := test.BuildApplicationEnv(db)

	aliasService := NewAliasService(appEnv)
	err := aliasService.Hide(test.UID, test.Alias)
	assert.NoError(t, err)
	aliases, err := aliasService.ListWithCounts(test.UID, "", "")

	assert.NoError(t, err)
	assert.Len(t, aliases.Items, 1)
	assert.Contains(t, aliases.Items, test.Alias2+" (1)")
}

func TestAliasService_ListWithCounts_grep(t *testing.T) {
	test.InsertTestData(db)

	aliasService := NewAliasService(test.BuildApplicationEnv(db))
	aliases, err := aliasService.ListWithCounts(test.UID, "b", "")

	assert.NoError(t, err)
	assert.Len(t, aliases.Items, 0)
}

func TestAliasService_ListBy(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestLocation(db, test.UID2)
	appEnv := test.BuildApplicationEnv(db)

	aliasService := NewAliasService(appEnv)
	aliases, err := aliasService.ListByFile(test.UID, &wizard.File{UniqueID: test.UniqueFileID})
	assert.NoError(t, err)
	assert.Len(t, aliases, 2)
	assert.Contains(t, aliases, test.Alias)
	assert.Contains(t, aliases, test.Alias2)

	aliases, err = aliasService.ListByText(test.UID2, &wizard.Txt{Value: test.Text})
	assert.NoError(t, err)
	assert.Len(t, aliases, 1)
	assert.Contains(t, aliases, test.Alias2)

	aliases, err = aliasService.ListByLocation(test.UID2, &wizard.LocData{
		Latitude:  test.Latitude,
		Longitude: test.Longitude,
	})
	assert.NoError(t, err)
	assert.Len(t, aliases, 1)
	assert.Contains(t, aliases, test.Alias2)
}
