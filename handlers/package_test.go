package handlers

import (
	"fmt"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageAction(t *testing.T) {
	test.InsertTestData(db)

	reqenv := test.BuildRequestEnv(db)
	packageService := repo.NewPackageService(reqenv)
	msg := buildMessage(test.UID3)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldName,
			Data: test.Package,
		},
		&wizard.Field{
			Name: FieldCreateOrDelete,
			Data: Create,
		},
		&wizard.Field{
			Name: FieldAliases,
			Data: test.Alias + "\n" + test.Alias2,
		},
	}
	packageAction(reqenv, msg, fields)

	packages, err := packageService.ListWithCounts(test.UID3)
	assert.NoError(t, err)
	assert.Len(t, packages, 1)
	assert.Contains(t, packages, fmt.Sprintf("%d@%s (2)", test.UID3, test.Package))

	fields.FindField(FieldCreateOrDelete).Data = Delete
	packageAction(reqenv, msg, fields)

	packages, err = packageService.ListWithCounts(test.UID3)
	assert.NoError(t, err)
	assert.Len(t, packages, 0)
}
