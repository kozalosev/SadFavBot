package handlers

import (
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsePackageName(t *testing.T) {
	pkgInfo, err := parsePackageName(test.PackageFullName)

	assert.NoError(t, err)
	assert.Equal(t, int64(test.UID), pkgInfo.UID)
	assert.Equal(t, test.Package, pkgInfo.Name)
}

func TestInstallPackageAction(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	aliasService := repo.NewAliasService(appenv)
	msg := buildMessage(test.UID3)
	fields := wizard.Fields{
		test.NewTextField(FieldName, test.PackageFullName),
		test.NewTextField(FieldConfirmation, No),
	}

	handler := NewInstallPackageHandler(appenv, nil)
	handler.installPackageAction(reqenv, msg, fields)

	aliases, err := aliasService.List(test.UID3)
	assert.NoError(t, err)
	assert.Len(t, aliases, 0)

	fields.FindField(FieldConfirmation).Data = wizard.Txt{Value: Yes}
	handler.installPackageAction(reqenv, msg, fields)

	aliasesPage, err := aliasService.ListWithCounts(test.UID3, "", "")
	assert.NoError(t, err)
	assert.Len(t, aliasesPage.Items, 1)
	assert.Contains(t, aliasesPage.Items, test.Alias2+" (1)")
}
