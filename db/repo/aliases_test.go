package repo

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testAlias = "test_alias"

func TestTrimCountSuffix(t *testing.T) {
	assert.Equal(t, testAlias, trimCountSuffix(testAlias+" (1)"))
	assert.Equal(t, testAlias, trimCountSuffix(testAlias))
	assert.Equal(t, testAlias+" (test)", trimCountSuffix(testAlias+" (test)"))
}

func TestAliasService_ListWithCounts(t *testing.T) {
	test.InsertTestData(db)

	aliasService := NewAliasService(test.BuildRequestEnv(db))
	aliases, err := aliasService.ListWithCounts(test.UID)

	assert.NoError(t, err)
	assert.Len(t, aliases, 2)
	assert.Contains(t, aliases, test.Alias+" (2)")
	assert.Contains(t, aliases, test.Alias2+" (1)")
}

func TestAliasService_List_noRows(t *testing.T) {
	test.InsertTestData(db)

	aliasService := NewAliasService(test.BuildRequestEnv(db))
	aliases, err := aliasService.List(test.UID - 1)

	assert.NoError(t, err)
	assert.Len(t, aliases, 0)
}
