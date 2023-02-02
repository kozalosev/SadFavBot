package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveFormAction(t *testing.T) {
	insertTestData(dbConn)

	reqenv := &base.RequestEnv{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: TestUID3},
		},
		Database: dbConn,
		Bot:      &base.BotAPI{DummyMode: true},
		Lang:     loc.NewPool("en").GetContext("en"),
	}
	fields := wizard.Fields{
		&wizard.Field{Name: "alias", Data: TestAlias},
		&wizard.Field{Name: "object", Type: TestType, Data: wizard.File{
			FileID:       TestFileID,
			FileUniqueID: TestUniqueFileID,
		}},
	}

	saveFormAction(reqenv, fields)

	countRes := dbConn.QueryRow("SELECT count(id) FROM item WHERE uid = $1", TestUID3)
	var count int
	err := countRes.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	itemsRes := dbConn.QueryRow("SELECT alias, type, file_id, file_unique_id FROM item WHERE uid = $1", TestUID3)
	var item queryResult
	err = itemsRes.Scan(&item.Name, &item.Type, &item.FileID, &item.FileUniqueID)
	assert.NoError(t, err)
	assert.Equal(t, TestAlias, item.Name)
	assert.Equal(t, TestType, item.Type)
	assert.Equal(t, TestFileID, item.FileID)
	assert.Equal(t, TestUniqueFileID, item.FileUniqueID)
}

type queryResult struct {
	Name         string
	Type         string
	FileID       string
	FileUniqueID string
}
