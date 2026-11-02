package handlers

import (
	"fmt"
	"testing"

	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPackageAction_MarkdownEscapedInCreationMessage(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := buildRequestEnvWithStrings(map[string]string{
		PackageStatusCreationSuccess: "Package: *%s*\n`/install %s`\n%s %s",
	})
	fields := wizard.Fields{
		test.NewTextField(FieldName, "kdr_flood"),
		test.NewTextField(FieldCreateOrDelete, Create),
		test.NewTextField(FieldAliases, test.Alias),
	}

	NewPackageHandler(appenv, nil).packageAction(reqenv, buildMessage(test.UID3), fields)

	output := appenv.Bot.(*base.FakeBotAPI).GetOutput().([]string)
	assert.Len(t, output, 1)
	assert.Contains(t, output[0], fmt.Sprintf("*%d@kdr\\_flood*", test.UID3))
	assert.Contains(t, output[0], fmt.Sprintf("`/install %d@kdr_flood`", test.UID3))
}

func TestPackageAction_Link(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := buildRequestEnvWithStrings(map[string]string{
		PackageStatusCreationSuccess: "Package: *%s*\n`/install %s`\n%s %s",
		PackageStatusNoRows:          "no rows",
	})
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		test.NewTextField(FieldName, test.Package),
		test.NewTextField(FieldCreateOrDelete, Link),
	}
	handler := NewPackageHandler(appenv, nil)
	handler.packageAction(reqenv, msg, fields)

	fields.FindField(FieldName).Data = wizard.Txt{Value: "no-such-package"}
	handler.packageAction(reqenv, msg, fields)

	output := appenv.Bot.(*base.FakeBotAPI).GetOutput().([]string)
	assert.Len(t, output, 2)
	assert.Contains(t, output[0], test.PackageUUID)
	assert.Contains(t, output[0], fmt.Sprintf("`/install %d@%s`", test.UID, test.Package))
	assert.Equal(t, "no rows", output[1])

	packages, err := repo.NewPackageService(appenv).ListWithCounts(test.UID, "")
	assert.NoError(t, err)
	assert.Contains(t, packages.Items, fmt.Sprintf("%d@%s (1)", test.UID, test.Package))
}

func TestNormalizePackageName(t *testing.T) {
	assert.Equal(t, "my-pack", normalizePackageName(test.UID, "my pack"))
	assert.Equal(t, "my-pack", normalizePackageName(test.UID, "my-pack"))
	assert.Equal(t, "my-pack", normalizePackageName(test.UID, fmt.Sprintf("%d@my pack", test.UID)))
	assert.Equal(t, "my-pack", normalizePackageName(test.UID, fmt.Sprintf("%d@my-pack", test.UID)))
	assert.Equal(t, fmt.Sprintf("%d@my-pack", test.UID2), normalizePackageName(test.UID, fmt.Sprintf("%d@my-pack", test.UID2)), "a foreign prefix must stay")
	assert.Empty(t, normalizePackageName(test.UID, fmt.Sprintf("%d@", test.UID)))
}

func TestPackageAction_FullNameIsEquivalentToShort(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := buildRequestEnvWithStrings(map[string]string{
		PackageStatusCreationSuccess: "Package: *%s*\n`/install %s`\n%s %s",
	})
	fields := wizard.Fields{
		test.NewTextField(FieldName, test.PackageFullName),
		test.NewTextField(FieldCreateOrDelete, Link),
	}

	NewPackageHandler(appenv, nil).packageAction(reqenv, buildMessage(test.UID), fields)

	output := appenv.Bot.(*base.FakeBotAPI).GetOutput().([]string)
	require.Len(t, output, 1)
	assert.Contains(t, output[0], test.PackageUUID)
}
