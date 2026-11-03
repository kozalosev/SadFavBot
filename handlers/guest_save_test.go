package handlers

import (
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/stretchr/testify/assert"
)

func TestGuestCommandsHandler_Save_CanHandle(t *testing.T) {
	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()

	for _, text := range []string{"save a", "Save a", "create a", "сохранить а", "Создать а", "save"} {
		assert.True(t, handler.CanHandle(reqenv, buildGuestMessage(test.UID3, mention+text)), text)
	}
	assert.False(t, handler.CanHandle(reqenv, buildGuestMessage(test.UID3, mention+"saved a")))
	assert.False(t, handler.CanHandle(reqenv, buildGuestMessage(test.UID3, mention+test.Alias)))
	assert.False(t, handler.CanHandle(reqenv, &tgbotapi.Message{Text: mention + "save a"}), "not a guest message")
}

func TestGuestCommandsHandler_Save_Handle(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()
	replyTo := &tgbotapi.Message{Sticker: &tgbotapi.Sticker{FileID: test.FileID, FileUniqueID: test.UniqueFileID}}

	// no alias
	result := handler.Handle(reqenv, buildGuestSaveMessage(mention+"save", replyTo))
	assert.Equal(t, reqenv.Lang.Tr(GuestSaveHintAliasTr), answerText(t, result))
	test.CheckRowsCount(t, db, 0, test.UID3, nil)

	// no reply
	result = handler.Handle(reqenv, buildGuestSaveMessage(mention+"save "+test.Alias, nil))
	assert.Equal(t, reqenv.Lang.Tr(GuestSaveHintReplyTr), answerText(t, result))
	test.CheckRowsCount(t, db, 0, test.UID3, nil)

	// forbidden symbols
	result = handler.Handle(reqenv, buildGuestSaveMessage(mention+"save a|b", replyTo))
	assert.Contains(t, answerText(t, result), "|")
	test.CheckRowsCount(t, db, 0, test.UID3, nil)

	// unsupported message
	result = handler.Handle(reqenv, buildGuestSaveMessage(mention+"save "+test.Alias, &tgbotapi.Message{}))
	assert.Equal(t, reqenv.Lang.Tr(GuestSaveUnsupportedTr), answerText(t, result))
	test.CheckRowsCount(t, db, 0, test.UID3, nil)

	// success; the alias may consist of several words
	result = handler.Handle(reqenv, buildGuestSaveMessage(mention+"Сохранить "+test.Alias+" two", replyTo))
	assert.Contains(t, answerText(t, result), test.Alias+" two")
	test.CheckRowsCount(t, db, 1, test.UID3, nil)

	// the command to use the fav is monospace, with the alias escaped
	result = handler.Handle(reqenv, buildGuestSaveMessage(mention+"save <b>`", replyTo))
	content := unwrapGuestResult(result).(tgbotapi.InlineQueryResultArticle).InputMessageContent.(tgbotapi.InputTextMessageContent)
	assert.Equal(t, tgbotapi.ModeHTML, content.ParseMode)
	assert.Contains(t, content.Text, "&lt;b&gt;`")
	test.CheckRowsCount(t, db, 2, test.UID3, nil)

	// the quotes around the alias are dropped, the mention may be at the end
	handler.Handle(reqenv, buildGuestSaveMessage(`сохрани «`+test.Alias+` three», `+mention, replyTo))
	unquoted := test.Alias + " three"
	test.CheckRowsCount(t, db, 1, test.UID3, &unquoted)

	// duplicate
	result = handler.Handle(reqenv, buildGuestSaveMessage(mention+"save "+test.Alias+" two", replyTo))
	assert.Equal(t, reqenv.Lang.Tr(SaveStatusDuplicate), answerText(t, result))
	test.CheckRowsCount(t, db, 3, test.UID3, nil)
}

func TestGuestCommandsHandler_Save_Handle_NewUser(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	msg := buildGuestMessage(999999, mention+"save "+test.Alias)
	msg.ReplyToMessage = &tgbotapi.Message{Text: test.Text}

	handler.Handle(test.BuildRequestEnv(), msg)
	test.CheckRowsCount(t, db, 1, 999999, nil)
}

func TestFavFromMessage(t *testing.T) {
	photo := favFromMessage(&tgbotapi.Message{
		Photo:   []tgbotapi.PhotoSize{{FileID: "small"}, {FileID: "big", FileUniqueID: "bigU"}},
		Caption: "caption",
	})
	assert.Equal(t, "big", photo.File.ID)
	assert.Equal(t, "caption", photo.File.Caption)

	text := favFromMessage(&tgbotapi.Message{Text: "text", Entities: []tgbotapi.MessageEntity{{Type: "bold"}}})
	assert.Equal(t, "text", text.Text.Value)
	assert.Len(t, text.Text.Entities, 1)

	assert.Equal(t, 1.5, favFromMessage(&tgbotapi.Message{Location: &tgbotapi.Location{Latitude: 1.5}}).Location.Latitude)
	assert.Nil(t, favFromMessage(&tgbotapi.Message{}))
}

func TestUnquote(t *testing.T) {
	assert.Equal(t, "alias", unquote(`"alias"`))
	assert.Equal(t, "alias two", unquote("«alias two»"))
	assert.Equal(t, "alias", unquote("„alias“"))
	assert.Equal(t, "alias", unquote("“alias”"))
	assert.Equal(t, "alias", unquote("'alias'"))
	assert.Equal(t, `"a" and "b"`, unquote(`"a" and "b"`), "the quotes don't enclose the whole text")
	assert.Equal(t, `"alias`, unquote(`"alias`))
	assert.Equal(t, `"`, unquote(`"`))
	assert.Equal(t, "alias", unquote("alias"))
}

func buildGuestSaveMessage(text string, replyTo *tgbotapi.Message) *tgbotapi.Message {
	msg := buildGuestMessage(test.UID3, text)
	msg.ReplyToMessage = replyTo
	return msg
}

func answerText(t *testing.T, result tgbotapi.InlineQueryResult) string {
	article, ok := unwrapGuestResult(result).(tgbotapi.InlineQueryResultArticle)
	assert.True(t, ok)
	return article.Title
}
