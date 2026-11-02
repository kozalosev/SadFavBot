package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/fp"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	showOutdated = "outdated"
	showNotOwner = "not owner"
	showEmpty    = "empty"
	showUsage    = "usage"
)

type fakeShowCache struct {
	mu      sync.Mutex
	entries map[string]*ShowEntry
}

func newFakeShowCache() *fakeShowCache {
	return &fakeShowCache{entries: make(map[string]*ShowEntry)}
}

func (cache *fakeShowCache) Put(_ context.Context, token string, entry *ShowEntry) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.entries[token] = entry
	return nil
}

func (cache *fakeShowCache) Get(_ context.Context, token string) (*ShowEntry, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if entry, ok := cache.entries[token]; ok {
		return entry, nil
	}
	return nil, ErrShowCacheMiss
}

func (cache *fakeShowCache) onlyToken(t *testing.T) string {
	require.Len(t, cache.entries, 1)
	return lo.Keys(cache.entries)[0]
}

func buildShowCommand(uid int64, args string) *tgbotapi.Message {
	msg := buildMessage(uid)
	msg.MessageID = 10
	msg.Text = "/show"
	if len(args) > 0 {
		msg.Text += " " + args
	}
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len("/show")}}
	return msg
}

func buildShowRequestEnv() *base.RequestEnv {
	return buildRequestEnvWithStrings(map[string]string{
		ShowStatusEmpty:    showEmpty,
		ShowStatusUsage:    showUsage,
		ShowStatusOutdated: showOutdated,
		ShowStatusNotOwner: showNotOwner,
	})
}

func favOf(t wizard.FieldType, id string) *dto.Fav {
	switch t {
	case wizard.Text:
		return &dto.Fav{ID: id, Type: t, Text: &wizard.Txt{Value: id}}
	case wizard.Location:
		return &dto.Fav{ID: id, Type: t, Location: &wizard.LocData{Latitude: 1, Longitude: 2}}
	default:
		return &dto.Fav{ID: id, Type: t, File: &wizard.File{ID: id, Caption: "caption"}}
	}
}

func TestSortFavsByType(t *testing.T) {
	favs := []*dto.Fav{
		favOf(wizard.Sticker, "sticker1"),
		favOf(wizard.Image, "image1"),
		favOf(wizard.Location, "location"),
		favOf(wizard.Text, "text"),
		favOf(wizard.Image, "image2"),
		favOf(wizard.Sticker, "sticker2"),
	}

	sortFavsByType(favs)

	assert.Equal(t, []string{"text", "image1", "image2", "sticker1", "sticker2", "location"}, favIDs(favs))
}

// requireOutput checks the number of requests sent by the bot and returns them.
func requireOutput(t *testing.T, bot *base.FakeBotAPI, n int) []tgbotapi.Chattable {
	output := requireAs[[]tgbotapi.Chattable](t, bot.GetOutput())
	require.Len(t, output, n)
	return output
}

// requireAs checks the type of the value and returns it converted.
func requireAs[T any](t *testing.T, v any) T {
	res, ok := v.(T)
	require.True(t, ok, "%T", v)
	return res
}

// requireAlert checks that the only request is the answer to the callback query with the alert.
func requireAlert(t *testing.T, bot *base.FakeBotAPI, expected string) {
	answer := requireAs[tgbotapi.CallbackConfig](t, requireOutput(t, bot, 1)[0])
	assert.True(t, answer.ShowAlert)
	assert.Equal(t, expected, answer.Text)
}

func favIDs(favs []*dto.Fav) []string {
	return fp.Map(favs, func(fav *dto.Fav) string { return fav.ID })
}

func TestCanEditInPlace(t *testing.T) {
	assert.True(t, canEditInPlace(favOf(wizard.Text, "a"), favOf(wizard.Text, "b")))
	assert.True(t, canEditInPlace(favOf(wizard.Image, "a"), favOf(wizard.Image, "b")))
	assert.True(t, canEditInPlace(favOf(wizard.Document, "a"), favOf(wizard.Document, "b")))
	assert.False(t, canEditInPlace(favOf(wizard.Image, "a"), favOf(wizard.Video, "b")))
	assert.False(t, canEditInPlace(favOf(wizard.Sticker, "a"), favOf(wizard.Sticker, "b")))
	assert.False(t, canEditInPlace(favOf(wizard.Voice, "a"), favOf(wizard.Voice, "b")))
	assert.False(t, canEditInPlace(favOf(wizard.Location, "a"), favOf(wizard.Location, "b")))
}

func TestShowEntry_JSON(t *testing.T) {
	entry := &ShowEntry{OwnerUID: test.UID, Favs: []*dto.Fav{
		favOf(wizard.Image, "image"),
		favOf(wizard.Text, "text"),
		favOf(wizard.Location, "location"),
	}}

	data, err := json.Marshal(entry)
	require.NoError(t, err)
	restored := new(ShowEntry)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, entry, restored)
}

func TestShowHandler_SeveralObjects(t *testing.T) {
	test.InsertTestData(db)
	appenv := test.BuildApplicationEnv(db)
	cache := newFakeShowCache()

	shown := testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeShow))
	NewShowHandler(appenv, cache).Handle(buildShowRequestEnv(), buildShowCommand(test.UID, test.Alias))
	assert.Equal(t, shown+1, testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeShow)))

	sticker := requireAs[tgbotapi.StickerConfig](t, requireOutput(t, appenv.Bot.(*base.FakeBotAPI), 1)[0])
	assert.Equal(t, 10, sticker.ReplyParameters.MessageID)

	token := cache.onlyToken(t)
	assert.Len(t, cache.entries[token].Favs, 2)
	assert.EqualValues(t, test.UID, cache.entries[token].OwnerUID)
	row := requireAs[tgbotapi.InlineKeyboardMarkup](t, sticker.ReplyMarkup).InlineKeyboard[0]
	assert.Equal(t, "1/2", row[1].Text)
	assert.Equal(t, "show:"+token+":1:0", *row[2].CallbackData)
	assert.Equal(t, "show:"+token+":1:0", *row[0].CallbackData) // wraps around
}

func TestShowHandler_SingleObjectHasNoKeyboard(t *testing.T) {
	test.InsertTestData(db)
	appenv := test.BuildApplicationEnv(db)
	cache := newFakeShowCache()

	shown := testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeShow))
	NewShowHandler(appenv, cache).Handle(buildShowRequestEnv(), buildShowCommand(test.UIDPhoto, test.AliasPhoto))
	assert.Equal(t, shown, testutil.ToFloat64(carouselCounter.WithLabelValues(metricModeShow)))

	photo := requireAs[tgbotapi.PhotoConfig](t, requireOutput(t, appenv.Bot.(*base.FakeBotAPI), 1)[0])
	assert.Nil(t, photo.ReplyMarkup)
	assert.Equal(t, test.CaptionPhoto, photo.Caption)
	assert.Empty(t, cache.entries)
}

func TestShowHandler_Replies(t *testing.T) {
	test.InsertTestData(db)

	tests := []struct {
		name     string
		args     string
		expected string
	}{
		{"unknown alias", "no-such-alias", showEmpty},
		{"no alias", "", showUsage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appenv := test.BuildApplicationEnv(db)
			NewShowHandler(appenv, newFakeShowCache()).Handle(buildShowRequestEnv(), buildShowCommand(test.UID, tt.args))

			assert.Equal(t, []string{tt.expected}, appenv.Bot.(*base.FakeBotAPI).GetOutput())
		})
	}
}

func TestShowHandler_NoUsageHintInGroups(t *testing.T) {
	appenv := test.BuildApplicationEnv(db)
	msg := buildShowCommand(test.UID, "")
	msg.Chat.Type = "supergroup"

	NewShowHandler(appenv, newFakeShowCache()).Handle(buildShowRequestEnv(), msg)

	assert.Empty(t, appenv.Bot.(*base.FakeBotAPI).GetOutput())
}

// buildShowCallback prepares a callback handler which has no access to the database.
func buildShowCallback(uid int64, data string) (*ShowCallbackHandler, *base.FakeBotAPI, *tgbotapi.CallbackQuery, *fakeShowCache) {
	appenv, bot, cache := buildShowCallbackEnv()
	query := buildCallbackQuery(uid, data)
	query.Message = &tgbotapi.Message{
		MessageID:      20,
		Chat:           tgbotapi.Chat{ID: 5, Type: "private"},
		ReplyToMessage: &tgbotapi.Message{MessageID: 10},
	}
	return NewShowCallbackHandler(appenv, cache), bot, query, cache
}

// buildShowCallbackEnv prepares the environment of a callback handler which has no access to the database.
func buildShowCallbackEnv() (*base.ApplicationEnv, *base.FakeBotAPI, *fakeShowCache) {
	bot := &base.FakeBotAPI{}
	return &base.ApplicationEnv{Bot: bot, Ctx: context.Background()}, bot, newFakeShowCache()
}

func buildCallbackQuery(uid int64, data string) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{ID: "query", From: &tgbotapi.User{ID: uid}, Data: data}
}

func putShowEntry(cache *fakeShowCache, favs ...*dto.Fav) {
	_ = cache.Put(context.Background(), "token", &ShowEntry{OwnerUID: test.UID, Favs: favs})
}

func TestShowCallbackHandler_EditsMessageOfTheSameType(t *testing.T) {
	handler, bot, query, cache := buildShowCallback(test.UID, "show:token:1:0")
	putShowEntry(cache, favOf(wizard.Image, "image1"), favOf(wizard.Image, "image2"))

	handler.Handle(buildShowRequestEnv(), query)

	output := requireOutput(t, bot, 2)
	edit := requireAs[tgbotapi.EditMessageMediaConfig](t, output[0])
	assert.Equal(t, 20, edit.MessageID)
	assert.Equal(t, "show:token:0:1", *edit.ReplyMarkup.InlineKeyboard[0][2].CallbackData)
	assert.IsType(t, tgbotapi.CallbackConfig{}, output[1])
}

func TestShowCallbackHandler_ResendsMessageOfAnotherType(t *testing.T) {
	handler, bot, query, cache := buildShowCallback(test.UID, "show:token:1:0")
	putShowEntry(cache, favOf(wizard.Image, "image"), favOf(wizard.Sticker, "sticker"))

	handler.Handle(buildShowRequestEnv(), query)

	output := requireOutput(t, bot, 3)
	assert.Equal(t, 20, requireAs[tgbotapi.DeleteMessageConfig](t, output[0]).MessageID)
	sticker := requireAs[tgbotapi.StickerConfig](t, output[1])
	assert.Equal(t, 10, sticker.ReplyParameters.MessageID)
	assert.NotNil(t, sticker.ReplyMarkup)
}

func TestShowCallbackHandler_ResendsStickers(t *testing.T) {
	handler, bot, query, cache := buildShowCallback(test.UID, "show:token:1:0")
	putShowEntry(cache, favOf(wizard.Sticker, "sticker1"), favOf(wizard.Sticker, "sticker2"))

	handler.Handle(buildShowRequestEnv(), query)

	output := requireOutput(t, bot, 3)
	assert.IsType(t, tgbotapi.DeleteMessageConfig{}, output[0])
	assert.IsType(t, tgbotapi.StickerConfig{}, output[1])
}

func TestShowCallbackHandler_WrapsAround(t *testing.T) {
	handler, bot, query, cache := buildShowCallback(test.UID, "show:token:-1:0")
	putShowEntry(cache, favOf(wizard.Text, "text1"), favOf(wizard.Text, "text2"), favOf(wizard.Text, "text3"))

	handler.Handle(buildShowRequestEnv(), query)

	edit := requireAs[tgbotapi.EditMessageTextConfig](t, requireOutput(t, bot, 2)[0])
	assert.Equal(t, "text3", edit.Text)
	assert.Equal(t, "3/3", edit.ReplyMarkup.InlineKeyboard[0][1].Text)
}

func TestShowCallbackHandler_Alerts(t *testing.T) {
	tests := []struct {
		name     string
		uid      int64
		data     string
		expected string
	}{
		{"foreign user", test.UID2, "show:token:1:0", showNotOwner},
		{"expired token", test.UID, "show:expired:1:0", showOutdated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, bot, query, cache := buildShowCallback(tt.uid, tt.data)
			putShowEntry(cache, favOf(wizard.Text, "text1"), favOf(wizard.Text, "text2"))

			handler.Handle(buildShowRequestEnv(), query)

			requireAlert(t, bot, tt.expected)
		})
	}
}

func TestShowCallbackHandler_IgnoresPressOnTheCounter(t *testing.T) {
	handler, bot, query, cache := buildShowCallback(test.UID, "show:token:0:0")
	putShowEntry(cache, favOf(wizard.Text, "text1"), favOf(wizard.Text, "text2"))

	handler.Handle(buildShowRequestEnv(), query)

	answer := requireAs[tgbotapi.CallbackConfig](t, requireOutput(t, bot, 1)[0])
	assert.False(t, answer.ShowAlert)
}

func TestBuildShowSendRequest_ByType(t *testing.T) {
	tests := map[wizard.FieldType]any{
		wizard.Text:      tgbotapi.MessageConfig{},
		wizard.Image:     tgbotapi.PhotoConfig{},
		wizard.Gif:       tgbotapi.AnimationConfig{},
		wizard.Video:     tgbotapi.VideoConfig{},
		wizard.Audio:     tgbotapi.AudioConfig{},
		wizard.Document:  tgbotapi.DocumentConfig{},
		wizard.Voice:     tgbotapi.VoiceConfig{},
		wizard.Sticker:   tgbotapi.StickerConfig{},
		wizard.VideoNote: tgbotapi.VideoNoteConfig{},
		wizard.Location:  tgbotapi.LocationConfig{},
	}
	for fieldType, expected := range tests {
		t.Run(strings.TrimSpace(string(fieldType)), func(t *testing.T) {
			request := buildShowSendRequest(5, 10, favOf(fieldType, "id"), nil)
			assert.IsType(t, expected, request)
		})
	}
}
