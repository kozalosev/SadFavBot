package handlers

import (
	"fmt"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageAction(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	packageService := repo.NewPackageService(appenv)
	msg := buildMessage(test.UID3)
	fields := wizard.Fields{
		test.NewTextField(FieldName, test.Package),
		test.NewTextField(FieldCreateOrDelete, Create),
		test.NewTextField(FieldAliases, test.Alias+"\n"+test.Alias2),
	}

	handler := NewPackageHandler(appenv, nil)
	handler.packageAction(reqenv, msg, fields)

	packages, err := packageService.ListWithCounts(test.UID3, "")
	assert.NoError(t, err)
	assert.Len(t, packages.Items, 1)
	assert.False(t, packages.HasNextPage)
	assert.Contains(t, packages.Items, fmt.Sprintf("%d@%s (2)", test.UID3, test.Package))
	assert.Equal(t, packages.GetLastItem(), fmt.Sprintf("%s (2)", test.Package))

	fields.FindField(FieldCreateOrDelete).Data = wizard.Txt{Value: Recreate}
	fields.FindField(FieldAliases).Data = wizard.Txt{Value: test.Alias}
	handler.packageAction(reqenv, msg, fields)

	packages, err = packageService.ListWithCounts(test.UID3, "")
	assert.NoError(t, err)
	assert.Len(t, packages.Items, 1)
	assert.Contains(t, packages.Items, fmt.Sprintf("%d@%s (1)", test.UID3, test.Package))
	assert.Equal(t, packages.GetLastItem(), fmt.Sprintf("%s (1)", test.Package))

	fields.FindField(FieldCreateOrDelete).Data = wizard.Txt{Value: Delete}
	handler.packageAction(reqenv, msg, fields)

	packages, err = packageService.ListWithCounts(test.UID3, "")
	assert.NoError(t, err)
	assert.Len(t, packages.Items, 0)
}
