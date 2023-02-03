package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"testing"
)

func TestDeleteFormAction(t *testing.T) {
	insertTestData(db)

	reqenv := buildRequestEnvDelete(TestUID)
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldDeleteAll, Data: No},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: wizard.File{UniqueID: TestUniqueFileID}},
	}

	deleteFormAction(reqenv, fields)

	testAlias := TestAlias
	checkRowsCount(t, 1, TestUID, &testAlias) // row with FileID_2 is on its place
	checkRowsCount(t, 2, TestUID, nil)        // rows with alias2 and alias+FileID_2

	fields.FindField(FieldDeleteAll).Data = Yes
	deleteFormAction(reqenv, fields)

	checkRowsCount(t, 0, TestUID, &testAlias)
}

func TestDeleteFormActionText(t *testing.T) {
	insertTestData(db)

	reqenv := buildRequestEnvDelete(TestUID3)
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldDeleteAll, Data: No},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: TestText},
	}

	deleteFormAction(reqenv, fields)

	checkRowsCount(t, 0, TestUID3, nil) // row with TestFileID is on its place
}

func buildRequestEnvDelete(uid int64) *base.RequestEnv {
	return &base.RequestEnv{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: uid},
		},
		Database: db,
		Bot:      &base.BotAPI{DummyMode: true},
		Lang:     loc.NewPool("en").GetContext("en"),
	}
}
