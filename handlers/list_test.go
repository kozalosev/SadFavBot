package handlers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchAliases(t *testing.T) {
	insertTestData(db)

	aliases, err := fetchAliases(db, TestUID)

	assert.NoError(t, err)
	assert.Len(t, aliases, 2)
	assert.Contains(t, aliases, TestAlias+" (2)")
	assert.Contains(t, aliases, TestAlias2+" (1)")
}

func TestFetchAliasesNoRows(t *testing.T) {
	insertTestData(db)

	aliases, err := fetchAliases(db, TestUID-1)

	assert.NoError(t, err)
	assert.Len(t, aliases, 0)
}
