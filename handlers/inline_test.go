package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestFindObjects(t *testing.T) {
	insertTestData(dbConn)

	reqenv := buildRequestEnv()
	objects := findObjects(reqenv)

	assert.Len(t, objects, 2)
	assert.Equal(t, TestFileID, objects[0].FileID)
	assert.Equal(t, TestFileID2, objects[1].FileID)
}

func TestMapper(t *testing.T) {
	insertTestData(dbConn)

	reqenv := buildRequestEnv()
	objects := findObjects(reqenv)

	inlineAnswer := generateMapper(nil)(objects[0])
	assert.Equal(t, "InlineQueryResultCachedSticker", reflect.TypeOf(inlineAnswer).Name())
}

func buildRequestEnv() *base.RequestEnv {
	return &base.RequestEnv{
		InlineQuery: &tgbotapi.InlineQuery{
			From:  &tgbotapi.User{ID: TestUID},
			Query: TestAlias,
		},
		Database: dbConn,
	}
}
