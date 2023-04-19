package repo

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageService_ListWithCounts(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	packageService := NewPackageService(test.BuildRequestEnv(db))
	packages, err := packageService.ListWithCounts(test.UID)

	assert.NoError(t, err)
	assert.Len(t, packages, 1)
	assert.Contains(t, packages, FormatPackageName(test.UID, test.Package)+" (1)")
}

func TestPackageService_Create(t *testing.T) {
	test.InsertTestData(db)

	packageService := NewPackageService(test.BuildRequestEnv(db))
	err := packageService.Create(test.UID, test.Package, []string{test.Alias2})
	assert.NoError(t, err)

	packages, err := packageService.ListWithCounts(test.UID)
	assert.NoError(t, err)
	assert.Len(t, packages, 1)
	assert.Contains(t, packages, FormatPackageName(test.UID, test.Package)+" (1)")

	var aliasID int
	err = db.QueryRow(ctx, "SELECT alias_id FROM package_aliases pa JOIN packages p ON p.id = pa.package_id WHERE p.name = $1", test.Package).Scan(&aliasID)
	assert.NoError(t, err)
	assert.Equal(t, test.Alias2ID, aliasID)
}

func TestPackageService_Delete(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	packageService := NewPackageService(test.BuildRequestEnv(db))
	err := packageService.Delete(test.UID, test.Package)
	assert.NoError(t, err)

	packages, err := packageService.ListWithCounts(test.UID)
	assert.NoError(t, err)
	assert.Len(t, packages, 0)
}
