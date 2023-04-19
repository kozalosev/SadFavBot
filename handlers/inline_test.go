package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMapper(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	reqenv := test.BuildRequestEnv(db)

	favsService := repo.NewFavsService(reqenv)
	objects, err := favsService.Find(query.From.ID, query.Query, false)
	assert.NoError(t, err)

	inlineAnswer := generateMapper(loc.NewPool("en").GetContext("en"))(objects[0])
	assert.Equal(t, "InlineQueryResultCachedSticker", reflect.TypeOf(inlineAnswer).Name())
}

func TestGetFavoritesInlineHandler_Handle(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	reqenv := test.BuildRequestEnv(db)

	handler := GetFavoritesInlineHandler{}
	assert.True(t, handler.CanHandle(query))
	handler.Handle(reqenv, query)

	bot := reqenv.Bot.(*base.FakeBotAPI)
	c := bot.GetOutput().(tgbotapi.InlineConfig)
	assert.Len(t, c.Results, 2)

	sticker1 := c.Results[0].(tgbotapi.InlineQueryResultCachedSticker)
	assert.NotEmpty(t, sticker1.ID)
	assert.Equal(t, string(test.Type), sticker1.Type)
	assert.Equal(t, test.FileID, sticker1.StickerID)

	sticker2 := c.Results[1].(tgbotapi.InlineQueryResultCachedSticker)
	assert.NotEmpty(t, sticker2.ID)
	assert.Equal(t, string(test.Type), sticker2.Type)
	assert.Equal(t, test.FileID2, sticker2.StickerID)
}
