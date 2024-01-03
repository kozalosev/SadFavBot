package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMapper(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	appenv := test.BuildApplicationEnv(db)

	favsService := repo.NewFavsService(appenv)
	objects, err := favsService.Find(query.From.ID, query.Query, false)
	assert.NoError(t, err)

	inlineAnswer := generateMapper(loc.NewPool("en").GetContext("en"))(objects[0])
	assert.Equal(t, "InlineQueryResultCachedSticker", reflect.TypeOf(inlineAnswer).Name())
}

func TestMapperForTextWithEntities(t *testing.T) {
	test.InsertTestData(db)
	_, err := db.Exec(ctx, "UPDATE texts SET entities = $2 WHERE id = $1", test.TextID, "[{\"type\": \"spoiler\", \"length\": 7, \"offset\": 0}]")
	assert.NoError(t, err)

	query := &tgbotapi.InlineQuery{
		From:  &tgbotapi.User{ID: test.UID2},
		Query: test.Alias2,
	}
	appenv := test.BuildApplicationEnv(db)

	favsService := repo.NewFavsService(appenv)
	objects, err := favsService.Find(query.From.ID, query.Query, false)
	assert.NoError(t, err)
	assert.Len(t, objects, 1)

	inlineAnswer := generateMapper(loc.NewPool("en").GetContext("en"))(objects[0])
	article, ok := inlineAnswer.(tgbotapi.InlineQueryResultArticle)
	assert.True(t, ok)
	content, ok := article.InputMessageContent.(tgbotapi.InputTextMessageContent)
	assert.True(t, ok)
	assert.NotNil(t, content.Entities)
}

func TestGetFavoritesInlineHandler_Handle(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()

	handler := NewGetFavoritesInlineHandler(appenv)
	assert.True(t, handler.CanHandle(reqenv, query))
	handler.Handle(reqenv, query)

	bot := appenv.Bot.(*base.FakeBotAPI)
	cc := bot.GetOutput().([]tgbotapi.Chattable)
	assert.Len(t, cc, 1)
	c := cc[0].(tgbotapi.InlineConfig)
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

func TestGetFavoritesInlineHandler_Handle_PhotoWithCaption(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	query.From.ID = test.UIDPhoto
	query.Query = test.AliasPhoto
	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()

	handler := NewGetFavoritesInlineHandler(appenv)
	assert.True(t, handler.CanHandle(reqenv, query))
	handler.Handle(reqenv, query)

	bot := appenv.Bot.(*base.FakeBotAPI)
	cc := bot.GetOutput().([]tgbotapi.Chattable)
	assert.Len(t, cc, 1)
	c := cc[0].(tgbotapi.InlineConfig)
	assert.Len(t, c.Results, 1)

	photo := c.Results[0].(tgbotapi.InlineQueryResultCachedPhoto)
	assert.NotEmpty(t, photo.ID)
	assert.Equal(t, "photo", photo.Type)
	assert.Equal(t, test.FileIDPhoto, photo.PhotoID)
	assert.Equal(t, test.CaptionPhoto, photo.Caption)
}
