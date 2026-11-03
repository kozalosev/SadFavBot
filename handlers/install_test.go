package handlers

import (
	"fmt"
	"testing"

	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
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

func TestSendCountOfAliasesInPackage_EscapedName(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := buildRequestEnvWithStrings(map[string]string{
		PackageStatusCreationSuccess: "created",
		PackageItems:                 "The package <i>%s</i> consists of:\n\n%s",
	})
	msg := buildMessage(test.UID3)
	fields := wizard.Fields{
		test.NewTextField(FieldName, "kdr_flood"),
		test.NewTextField(FieldCreateOrDelete, Create),
		test.NewTextField(FieldAliases, test.Alias),
	}
	NewPackageHandler(appenv, nil).packageAction(reqenv, msg, fields)

	name := fmt.Sprintf("%d@kdr_flood", test.UID3)
	handler := NewInstallPackageHandler(appenv, nil)
	sendCountOfAliasesInPackage(handler, reqenv, msg, name)

	output := appenv.Bot.(*base.FakeBotAPI).GetOutput().([]string)
	assert.Len(t, output, 2)
	// the underscore can't be escaped inside an entity of legacy Markdown (#103), and needn't be in HTML
	assert.Contains(t, output[1], fmt.Sprintf("<i>%d@kdr_flood</i>", test.UID3))
}
