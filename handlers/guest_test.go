package handlers

import (
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/stretchr/testify/assert"
)

const mention = "@TestMockBotAPI "

func TestGuestFavoritesHandler_CanHandle(t *testing.T) {
	handler := NewGuestFavoritesHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()

	assert.True(t, handler.CanHandle(reqenv, buildGuestMessage(test.UID, mention+test.Alias)))
	assert.False(t, handler.CanHandle(reqenv, &tgbotapi.Message{Text: test.Alias}))
}

func TestGuestFavoritesHandler_Handle(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestFavoritesHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()

	result := handler.Handle(reqenv, buildGuestMessage(test.UID, mention+test.Alias))
	assert.IsType(t, tgbotapi.InlineQueryResultCachedSticker{}, result)
	assert.Equal(t, test.FileID, result.(tgbotapi.InlineQueryResultCachedSticker).StickerID)
}

func TestGuestFavoritesHandler_Handle_CallerFallback(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestFavoritesHandler(test.BuildApplicationEnv(db))
	msg := buildGuestMessage(test.UID, mention+test.Alias)
	msg.GuestBotCallerUser, msg.From = msg.From, nil

	result := handler.Handle(test.BuildRequestEnv(), msg)
	assert.IsType(t, tgbotapi.InlineQueryResultCachedSticker{}, result)
}

func TestGuestFavoritesHandler_Handle_NothingToAnswer(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestFavoritesHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()

	// unknown alias
	result := handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"unknown"))
	assert.IsType(t, tgbotapi.InlineQueryResultArticle{}, result)
	assert.Contains(t, result.(tgbotapi.InlineQueryResultArticle).Title, "unknown")

	// only the mention
	result = handler.Handle(reqenv, buildGuestMessage(test.UID, "@TestMockBotAPI"))
	assert.IsType(t, tgbotapi.InlineQueryResultArticle{}, result)
	assert.Equal(t, reqenv.Lang.Tr(GuestEmptyQueryTr), result.(tgbotapi.InlineQueryResultArticle).Title)

	// no user at all
	assert.Nil(t, handler.Handle(reqenv, &tgbotapi.Message{GuestQueryID: "guest-query", Text: mention + test.Alias}))
}

func TestStripBotMention(t *testing.T) {
	assert.Equal(t, "my alias", stripBotMention("@SadFavBot my alias", newBotMentionRegexp("SadFavBot")))
	assert.Equal(t, "my alias", stripBotMention("my alias @sadfavbot", newBotMentionRegexp("SadFavBot")))
	assert.Equal(t, "my alias", stripBotMention("  @SadFavBot  my alias ", newBotMentionRegexp("SadFavBot")))
	assert.Equal(t, "@SadFavBotNot x", stripBotMention("@SadFavBotNot x", newBotMentionRegexp("SadFavBot")))
	assert.Equal(t, "my alias", stripBotMention("my alias, @SadFavBot", newBotMentionRegexp("SadFavBot")))
	assert.Equal(t, "my alias", stripBotMention("@SadFavBot, my alias", newBotMentionRegexp("SadFavBot")))
	assert.Equal(t, "a, b", stripBotMention("a, b,@SadFavBot", newBotMentionRegexp("SadFavBot")))
	assert.Empty(t, stripBotMention("@SadFavBot", newBotMentionRegexp("SadFavBot")))
}

func buildGuestMessage(uid int64, text string) *tgbotapi.Message {
	return &tgbotapi.Message{
		From:         &tgbotapi.User{ID: uid},
		Text:         text,
		GuestQueryID: "guest-query",
	}
}
