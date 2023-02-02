package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteFormAction(t *testing.T) {
	insertTestData(dbConn)

	reqenv := &base.RequestEnv{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: TestUID},
		},
		Database: dbConn,
		Bot:      &base.BotAPI{DummyMode: true},
		Lang:     loc.NewPool("en").GetContext("en"),
	}
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldDeleteAll, Data: No},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: wizard.File{FileUniqueID: TestUniqueFileID}},
	}

	deleteFormAction(reqenv, fields)

	countRes := dbConn.QueryRow("SELECT count(id) FROM item WHERE uid = $1 AND alias = $2", TestUID, TestAlias)
	var count int
	err := countRes.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count) // row with FileID_2 is on its place

	countRes = dbConn.QueryRow("SELECT count(id) FROM item WHERE uid = $1", TestUID)
	err = countRes.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count) // rows with alias2 and alias+FileID_2

	fields.FindField(FieldDeleteAll).Data = Yes
	deleteFormAction(reqenv, fields)

	countRes = dbConn.QueryRow("SELECT count(id) FROM item WHERE uid = $1 AND alias = $2", TestUID, TestAlias)
	err = countRes.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	countRes = dbConn.QueryRow("SELECT count(id) FROM item WHERE uid = $1", TestUID)
	err = countRes.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
