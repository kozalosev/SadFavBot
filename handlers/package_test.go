package handlers

import (
	"fmt"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreatePackage(t *testing.T) {
	insertTestData(db)

	err := createPackage(ctx, db, TestUID, TestPackage, []string{TestAlias2})
	assert.NoError(t, err)

	packages, err := fetchPackages(db, TestUID)
	assert.NoError(t, err)
	assert.Len(t, packages, 1)
	assert.Contains(t, packages, formatPackageName(TestUID, TestPackage)+" (1)")

	var aliasID int
	err = db.QueryRow("SELECT alias_id FROM package_aliases pa JOIN packages p ON p.id = pa.package_id WHERE p.name = $1", TestPackage).Scan(&aliasID)
	assert.NoError(t, err)
	assert.Equal(t, TestAlias2ID, aliasID)
}

func TestDeletePackage(t *testing.T) {
	insertTestData(db)
	insertTestPackages(db)

	err := deletePackage(db, TestUID, TestPackage)
	assert.NoError(t, err)

	packages, err := fetchPackages(db, TestUID)
	assert.NoError(t, err)
	assert.Len(t, packages, 0)
}

func TestPackageAction(t *testing.T) {
	insertTestData(db)

	reqenv := buildRequestEnv()
	msg := buildMessage(TestUID3)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldName,
			Data: TestPackage,
		},
		&wizard.Field{
			Name: FieldCreateOrDelete,
			Data: Create,
		},
		&wizard.Field{
			Name: FieldAliases,
			Data: TestAlias + "\n" + TestAlias2,
		},
	}
	packageAction(reqenv, msg, fields)

	packages, err := fetchPackages(db, TestUID3)
	assert.NoError(t, err)
	assert.Len(t, packages, 1)
	assert.Contains(t, packages, fmt.Sprintf("%d@%s (2)", TestUID3, TestPackage))

	fields.FindField(FieldCreateOrDelete).Data = Delete
	packageAction(reqenv, msg, fields)

	packages, err = fetchPackages(db, TestUID3)
	assert.NoError(t, err)
	assert.Len(t, packages, 0)
}
