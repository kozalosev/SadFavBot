package repo

import (
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"testing"
)

var (
	packageInfo = &PackageInfo{
		UID:  test.UID,
		Name: test.Package,
	}
)

func TestResolveName(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	packageService := NewPackageService(test.BuildApplicationEnv(db))
	packageName, err := packageService.ResolveName(test.PackageUUID)

	assert.NoError(t, err)
	assert.Equal(t, test.PackageFullName, packageName)
}

func TestFetchCountOfAliasesInPackage(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	packageService := NewPackageService(test.BuildApplicationEnv(db))
	aliases, err := packageService.ListAliases(packageInfo)

	assert.NoError(t, err)
	assert.Len(t, aliases, 1)
	assert.Contains(t, aliases, test.Alias2)
}

func TestInstallPackage(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	packageService := NewPackageService(test.BuildApplicationEnv(db))
	installed, err := packageService.Install(test.UID3, packageInfo)

	assert.NoError(t, err)
	assert.Len(t, installed, 1)
	assert.Contains(t, installed, test.Alias2)

	res, err := db.Query(ctx, "SELECT DISTINCT alias_id FROM favs WHERE uid = $1", test.UID3)
	assert.NoError(t, err)
	var (
		arr  []int
		elem int
	)
	for res.Next() {
		assert.NoError(t, res.Scan(&elem))
		arr = append(arr, elem)
	}
	assert.Len(t, arr, 1)
	assert.Contains(t, arr, test.Alias2ID)
}

func TestInstallPackageWithMoreAliases(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	_, err := db.Exec(ctx, "INSERT INTO package_aliases(package_id, alias_id) VALUES ($1, $2)", test.PackageID, test.AliasID)
	assert.NoError(t, err)

	packageService := NewPackageService(test.BuildApplicationEnv(db))
	installed, err := packageService.Install(test.UID3, packageInfo)

	assert.NoError(t, err)
	assert.Len(t, installed, 2)
	assert.Contains(t, installed, test.Alias)
	assert.Contains(t, installed, test.Alias2)
}

func TestInstallPackageWithLink(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	_, err := db.Exec(ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = $2", test.UID, test.AliasID)
	assert.NoError(t, err)
	_, err = db.Exec(ctx, "INSERT INTO links(uid, alias_id, linked_alias_id) VALUES ($1, $2, $3)", test.UID, test.AliasID, test.Alias2ID)
	assert.NoError(t, err)
	_, err = db.Exec(ctx, "INSERT INTO package_aliases(package_id, alias_id) VALUES ($1, $2)", test.PackageID, test.AliasID)
	assert.NoError(t, err)

	packageService := NewPackageService(test.BuildApplicationEnv(db))
	installed, err := packageService.Install(test.UID3, packageInfo)

	assert.NoError(t, err)
	assert.Len(t, installed, 2)
	assert.Contains(t, installed, test.Alias)
	assert.Contains(t, installed, test.Alias2)
}

func TestInstallPackageWithLocation(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)
	test.InsertTestLocation(db, test.UID)

	appEnv := test.BuildApplicationEnv(db)
	packageService := NewPackageService(appEnv)
	installed, err := packageService.Install(test.UID3, packageInfo)

	assert.NoError(t, err)
	assert.Len(t, installed, 1)
	assert.Contains(t, installed, test.Alias2)

	favsService := NewFavsService(appEnv)
	favs, err := favsService.Find(test.UID3, test.Alias2, false, 0)
	assert.NoError(t, err)
	assert.Len(t, favs, 2)
	locFavs := funk.Filter(favs, func(f *dto.Fav) bool {
		return f.Location != nil
	})
	assert.Len(t, locFavs, 1)
}

func TestRemoveDuplicates(t *testing.T) {
	arr := removeDuplicates([]int{6, 6, 6})

	assert.Len(t, arr, 1)
	assert.Contains(t, arr, 6)
}
