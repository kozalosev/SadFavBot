package handlers

import (
	"context"
	"encoding/json"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const guestInlineMessageID = "inline-message"

func TestGuestFavoritesHandler_Handle_Carousel(t *testing.T) {
	test.InsertTestData(db)
	_, err := db.Exec(context.Background(), "INSERT INTO favs(uid, type, alias_id, file_id, file_unique_id) VALUES ($1, 'image', $2, 'photo2', 'photo2')",
		test.UIDPhoto, test.AliasPhotoID)
	require.NoError(t, err)

	cache := newFakeShowCache()
	handler := NewGuestFavoritesHandler(test.BuildApplicationEnv(db), cache)

	shown := testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeGuest))
	result := handler.Handle(test.BuildRequestEnv(), buildGuestMessage(test.UIDPhoto, mention+test.AliasPhoto))
	assert.Equal(t, shown+1, testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeGuest)))
	photo := requireAs[tgbotapi.InlineQueryResultCachedPhoto](t, result)
	require.NotNil(t, photo.ReplyMarkup)

	token := cache.onlyToken(t)
	row := photo.ReplyMarkup.InlineKeyboard[0]
	assert.Equal(t, "1/2", row[1].Text)
	assert.Equal(t, "gshow:"+token+":1:0", *row[2].CallbackData)
	assert.Equal(t, int64(test.UIDPhoto), cache.entries[token].OwnerUID)
}

func TestGuestFavoritesHandler_Handle_NoCarouselForStickers(t *testing.T) {
	test.InsertTestData(db)

	cache := newFakeShowCache()
	handler := NewGuestFavoritesHandler(test.BuildApplicationEnv(db), cache)

	// there are two stickers, but they can't replace each other in an inline message
	shown := testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeGuest))
	result := handler.Handle(test.BuildRequestEnv(), buildGuestMessage(test.UID, mention+test.Alias))
	assert.IsType(t, tgbotapi.InlineQueryResultCachedSticker{}, result)
	assert.Empty(t, cache.entries)
	assert.Equal(t, shown, testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeGuest)))
}

func TestFilterGuestCarousel(t *testing.T) {
	favs := []*dto.Fav{
		favOf(wizard.Image, "image"),
		favOf(wizard.Text, "text"),
		favOf(wizard.Sticker, "sticker"),
		favOf(wizard.Video, "video"),
		favOf(wizard.Document, "document"),
	}
	assert.Equal(t, []string{"image", "video", "document"}, favIDs(filterGuestCarousel(favs)))

	favs = []*dto.Fav{favOf(wizard.Text, "text1"), favOf(wizard.Image, "image"), favOf(wizard.Text, "text2")}
	assert.Equal(t, []string{"text1", "text2"}, favIDs(filterGuestCarousel(favs)))

	favs = []*dto.Fav{favOf(wizard.Sticker, "sticker1"), favOf(wizard.Sticker, "sticker2")}
	assert.Equal(t, []string{"sticker1"}, favIDs(filterGuestCarousel(favs)))
}

func TestWithReplyMarkup(t *testing.T) {
	keyboard := buildCarouselKeyboard(guestShowCallbackPrefix, "token", 0, 2)
	mapper := generateMapper(test.BuildRequestEnv().Lang)
	for _, fieldType := range []wizard.FieldType{wizard.Text, wizard.Image, wizard.Gif, wizard.Video, wizard.Audio, wizard.Document} {
		t.Run(string(fieldType), func(t *testing.T) {
			data, err := json.Marshal(withReplyMarkup(mapper(favOf(fieldType, "id")), keyboard))
			require.NoError(t, err)
			assert.Contains(t, string(data), `"reply_markup":{"inline_keyboard":`)
		})
	}
}

// buildGuestShowCallback prepares a callback handler which has no access to the database.
func buildGuestShowCallback(uid int64, data string) (*GuestShowCallbackHandler, *base.FakeBotAPI, *tgbotapi.CallbackQuery, *fakeShowCache) {
	appenv, bot, cache := buildShowCallbackEnv()
	query := buildCallbackQuery(uid, data)
	query.InlineMessageID = guestInlineMessageID
	return NewGuestShowCallbackHandler(appenv, cache), bot, query, cache
}

func TestGuestShowCallbackHandler_EditsInlineMessage(t *testing.T) {
	handler, bot, query, cache := buildGuestShowCallback(test.UID, "gshow:token:1:0")
	putShowEntry(cache, favOf(wizard.Image, "image"), favOf(wizard.Video, "video"))

	handler.Handle(buildShowRequestEnv(), query)

	output := requireOutput(t, bot, 2)
	edit := requireAs[tgbotapi.EditMessageMediaConfig](t, output[0])
	assert.Equal(t, guestInlineMessageID, edit.InlineMessageID)
	assert.IsType(t, &tgbotapi.InputMediaVideo{}, edit.Media)
	assert.Equal(t, "gshow:token:0:1", *edit.ReplyMarkup.InlineKeyboard[0][2].CallbackData)
	assert.IsType(t, tgbotapi.CallbackConfig{}, output[1])
}

func TestGuestShowCallbackHandler_EditsText(t *testing.T) {
	handler, bot, query, cache := buildGuestShowCallback(test.UID, "gshow:token:-1:0")
	putShowEntry(cache, favOf(wizard.Text, "text1"), favOf(wizard.Text, "text2"), favOf(wizard.Text, "text3"))

	handler.Handle(buildShowRequestEnv(), query)

	edit := requireAs[tgbotapi.EditMessageTextConfig](t, requireOutput(t, bot, 2)[0])
	assert.Equal(t, guestInlineMessageID, edit.InlineMessageID)
	assert.Equal(t, "text3", edit.Text)
	assert.Equal(t, "3/3", edit.ReplyMarkup.InlineKeyboard[0][1].Text)
}

func TestGuestShowCallbackHandler_NotOwner(t *testing.T) {
	handler, bot, query, cache := buildGuestShowCallback(test.UID2, "gshow:token:1:0")
	putShowEntry(cache, favOf(wizard.Text, "text1"), favOf(wizard.Text, "text2"))

	handler.Handle(buildShowRequestEnv(), query)

	requireAlert(t, bot, showNotOwner)
}
