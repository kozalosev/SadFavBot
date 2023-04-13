package handlers

import (
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchCountOfAliasesInPackage(t *testing.T) {
	insertTestData(db)
	insertTestPackages(db)

	aliases, err := fetchAliasesInPackage(ctx, db, TestPackageFullName)
	assert.NoError(t, err)
	assert.Len(t, aliases, 1)
	assert.Contains(t, aliases, TestAlias2)
}

func TestInstallPackage(t *testing.T) {
	insertTestData(db)
	insertTestPackages(db)

	installed, err := installPackage(ctx, db, TestUID3, TestPackageFullName)
	assert.NoError(t, err)
	assert.Len(t, installed, 1)
	assert.Contains(t, installed, TestAlias2)

	res, err := db.Query("SELECT DISTINCT alias_id FROM favs WHERE uid = $1", TestUID3)
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
	assert.Contains(t, arr, TestAlias2ID)
}

func TestInstallPackageWithMoreAliases(t *testing.T) {
	insertTestData(db)
	insertTestPackages(db)

	_, err := db.Exec("INSERT INTO package_aliases(package_id, alias_id) VALUES ($1, $2)", TestPackageID, TestAliasID)
	assert.NoError(t, err)

	installed, err := installPackage(ctx, db, TestUID3, TestPackageFullName)
	assert.NoError(t, err)
	assert.Len(t, installed, 2)
	assert.Contains(t, installed, TestAlias)
	assert.Contains(t, installed, TestAlias2)
}

func TestInstallPackageWithLink(t *testing.T) {
	insertTestData(db)
	insertTestPackages(db)

	_, err := db.Exec("DELETE FROM favs WHERE uid = $1 AND alias_id = $2", TestUID, TestAliasID)
	assert.NoError(t, err)
	_, err = db.Exec("INSERT INTO links(uid, alias_id, linked_alias_id) VALUES ($1, $2, $3)", TestUID, TestAliasID, TestAlias2ID)
	assert.NoError(t, err)
	_, err = db.Exec("INSERT INTO package_aliases(package_id, alias_id) VALUES ($1, $2)", TestPackageID, TestAliasID)
	assert.NoError(t, err)

	installed, err := installPackage(ctx, db, TestUID3, TestPackageFullName)
	assert.NoError(t, err)
	assert.Len(t, installed, 2)
	assert.Contains(t, installed, TestAlias)
	assert.Contains(t, installed, TestAlias2)
}

func TestParsePackageName(t *testing.T) {
	pkgInfo, err := parsePackageName(TestPackageFullName)

	assert.NoError(t, err)
	assert.Equal(t, int64(TestUID), pkgInfo.uid)
	assert.Equal(t, TestPackage, pkgInfo.name)
}

func TestRemoveDuplicates(t *testing.T) {
	arr := removeDuplicates([]int{6, 6, 6})

	assert.Len(t, arr, 1)
	assert.Contains(t, arr, 6)
}

func TestInstallPackageAction(t *testing.T) {
	insertTestData(db)
	insertTestPackages(db)

	reqenv := buildRequestEnv()
	msg := buildMessage(TestUID3)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldName,
			Data: TestPackageFullName,
		},
		&wizard.Field{
			Name: FieldConfirmation,
			Data: No,
		},
	}
	installPackageAction(reqenv, msg, fields)

	aliases, err := fetchAliases(ctx, db, TestUID3)
	assert.NoError(t, err)
	assert.Len(t, aliases, 0)

	fields.FindField(FieldConfirmation).Data = Yes
	installPackageAction(reqenv, msg, fields)

	aliases, err = fetchAliases(ctx, db, TestUID3)
	assert.NoError(t, err)
	assert.Len(t, aliases, 1)
	assert.Contains(t, aliases, TestAlias2+" (1)")
}
