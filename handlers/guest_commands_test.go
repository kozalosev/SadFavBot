package handlers

import (
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/stretchr/testify/assert"
)

func TestParseGuestCommand(t *testing.T) {
	cases := []struct {
		text string
		cmd  guestCommand
		args string
	}{
		{"save a b", guestCmdSave, "a b"},
		{"Create a", guestCmdSave, "a"},
		{"  Сохранить   a  b ", guestCmdSave, "a  b"},
		{"создать a", guestCmdSave, "a"},
		{"delete a", guestCmdDelete, "a"},
		{"remove a", guestCmdDelete, "a"},
		{"удали a", guestCmdDelete, "a"},
		{"Удалить a", guestCmdDelete, "a"},
		{"ref", guestCmdRef, ""},
		{"reference", guestCmdRef, ""},
		{"реф", guestCmdRef, ""},
		{"info a", guestCmdRef, "a"},
		{"Инфо", guestCmdRef, ""},
		{"информация", guestCmdRef, ""},
		{"list", guestCmdList, ""},
		{"список foo", guestCmdList, "foo"},
		{"покажи все foo", guestCmdList, "foo"},
		{"link a -> b", guestCmdLink, "a -> b"},
		{"ссылка a -> b", guestCmdLink, "a -> b"},
		{"алиас a -> b", guestCmdLink, "a -> b"},
		{"псевдоним a -> b", guestCmdLink, "a -> b"},
		{"rmlink a", guestCmdRmLink, "a"},
		{"remove link a", guestCmdRmLink, "a"},
		{"Delete  Link a", guestCmdRmLink, "a"},
		{"удалить ссылку a", guestCmdRmLink, "a"},
		{"удалить алиас a", guestCmdRmLink, "a"},
		{"удалить псевдоним a", guestCmdRmLink, "a"},
		{"help", guestCmdHelp, ""},
		{"помощь", guestCmdHelp, ""},
		{"справка", guestCmdHelp, ""},
		{"saved a", 0, ""},
		{"покажи", 0, ""},
		{test.Alias, 0, ""},
		{"", 0, ""},
	}
	for _, c := range cases {
		cmd, args := parseGuestCommand(c.text)
		assert.Equal(t, c.cmd, cmd, c.text)
		assert.Equal(t, c.args, args, c.text)
	}
}

func TestGuestCommandsHandler_CanHandle(t *testing.T) {
	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()

	for _, text := range []string{"save a", "удалить a", "info", "список", "ссылка a -> b", "help"} {
		assert.True(t, handler.CanHandle(reqenv, buildGuestMessage(test.UID, mention+text)), text)
	}
	assert.False(t, handler.CanHandle(reqenv, buildGuestMessage(test.UID, mention+test.Alias)))
	assert.False(t, handler.CanHandle(reqenv, &tgbotapi.Message{Text: mention + "help"}), "not a guest message")
}

func TestGuestCommandsHandler_Delete(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()
	lc := reqenv.Lang

	result := handler.Handle(reqenv, buildGuestReply(mention+"delete", stickerMessage()))
	assert.Equal(t, lc.Tr(GuestDeleteHintAliasTr), answerMessage(t, result))
	test.CheckRowsCount(t, db, 3, test.UID, nil)

	result = handler.Handle(reqenv, buildGuestReply(mention+"delete "+test.Alias, nil))
	assert.Equal(t, lc.Tr(GuestHintReplyTr), answerMessage(t, result))
	test.CheckRowsCount(t, db, 3, test.UID, nil)

	result = handler.Handle(reqenv, buildGuestReply(mention+"delete "+test.Alias, stickerMessage()))
	assert.Equal(t, lc.Tr(DeleteStatusSuccess), answerMessage(t, result))
	test.CheckRowsCount(t, db, 2, test.UID, nil)

	result = handler.Handle(reqenv, buildGuestReply(mention+"удалить "+test.Alias, stickerMessage()))
	assert.Equal(t, lc.Tr(DeleteStatusNoRows), answerMessage(t, result))
	test.CheckRowsCount(t, db, 2, test.UID, nil)

	result = handler.Handle(reqenv, buildGuestReply(mention+"delete "+test.Alias, &tgbotapi.Message{}))
	assert.Equal(t, lc.Tr(DeleteStatusNoRows), answerMessage(t, result))
	test.CheckRowsCount(t, db, 2, test.UID, nil)
}

func TestGuestCommandsHandler_Ref(t *testing.T) {
	test.InsertTestData(db)
	test.InsertTestPackages(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()
	lc := reqenv.Lang

	// by the object
	text := answerMessage(t, handler.Handle(reqenv, buildGuestReply(mention+"info", stickerMessage())))
	assert.Contains(t, text, lc.Tr(RefStatusSuccess))
	assert.Contains(t, text, LinePrefix+test.Alias)
	assert.Contains(t, text, LinePrefix+test.Alias2)
	assert.Contains(t, text, lc.Tr(RefStatusSuccessWithPackages))

	unknown := &tgbotapi.Message{Sticker: &tgbotapi.Sticker{FileID: "unknown", FileUniqueID: "unknown"}}
	result := handler.Handle(reqenv, buildGuestReply(mention+"ref", unknown))
	assert.Equal(t, lc.Tr(RefStatusNoFavs), answerMessage(t, result))

	// by the alias, without the reply
	text = answerMessage(t, handler.Handle(reqenv, buildGuestReply(mention+"ref "+test.Alias2, nil)))
	assert.Contains(t, text, lc.Tr(RefStatusSuccessPackagesByAlias))
	result = handler.Handle(reqenv, buildGuestReply(mention+"ref "+test.Alias, nil))
	assert.Equal(t, lc.Tr(RefStatusNoPacks), answerMessage(t, result))

	result = handler.Handle(reqenv, buildGuestReply(mention+"ref", nil))
	assert.Equal(t, lc.Tr(GuestHintReplyTr), answerMessage(t, result))
}

func TestGuestCommandsHandler_List(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()
	lc := reqenv.Lang

	text := answerMessage(t, handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"list")))
	assert.Contains(t, text, lc.Tr(ListStatusSuccessFavs))
	assert.Contains(t, text, test.Alias)
	assert.Contains(t, text, test.Alias2)

	text = answerMessage(t, handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"покажи все "+test.Alias2)))
	assert.Contains(t, text, test.Alias2)

	result := handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"list unknown"))
	assert.Equal(t, lc.Tr(ListStatusNoRowsFavs), answerMessage(t, result))
}

func TestGuestCommandsHandler_Link(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()
	lc := reqenv.Lang

	for _, text := range []string{"link", "link foo", "link foo ->", "link -> " + test.Alias} {
		result := handler.Handle(reqenv, buildGuestMessage(test.UID, mention+text))
		assert.Equal(t, lc.Tr(GuestLinkHintTr), answerMessage(t, result), text)
	}

	result := handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"link a|b -> "+test.Alias))
	assert.Contains(t, answerMessage(t, result), "|")

	result = handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"ссылка foo bar -> unknown"))
	assert.Equal(t, lc.Tr(LinkStatusNoAlias), answerMessage(t, result))

	result = handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"ссылка foo bar → "+test.Alias))
	assert.Equal(t, lc.Tr(LinkStatusSuccess), answerMessage(t, result))
	links, err := fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo bar -> " + test.Alias}, links)

	result = handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"link foo bar -> "+test.Alias))
	assert.Equal(t, lc.Tr(LinkStatusDuplicate), answerMessage(t, result))

	result = handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"link "+test.Alias2+" -> "+test.Alias))
	assert.Equal(t, lc.Tr(LinkStatusDuplicateFav), answerMessage(t, result))
}

func TestGuestCommandsHandler_RmLink(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()
	lc := reqenv.Lang

	assert.NoError(t, handler.linkService.Create(test.UID, "foo bar", test.Alias))

	result := handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"remove link"))
	assert.Equal(t, lc.Tr(GuestRmLinkHintTr), answerMessage(t, result))

	result = handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"удалить ссылку foo bar"))
	assert.Equal(t, lc.Tr(RmLinkStatusSuccess), answerMessage(t, result))
	links, err := fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Empty(t, links)

	result = handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"rmlink foo bar"))
	assert.Equal(t, lc.Tr(RmLinkStatusNoRows), answerMessage(t, result))
}

func TestGuestCommandsHandler_ListPlaceholder(t *testing.T) {
	test.InsertTestData(db)

	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()

	list, ok := handler.Handle(reqenv, buildGuestReply(mention+"list", nil)).(base.PlaceholderGuestResult)
	assert.True(t, ok)
	assert.Equal(t, reqenv.Lang.Tr(GuestListPlaceholderTr), list.Placeholder)

	// kept as is
	for _, text := range []string{"ref", "help", "save", "delete " + test.Alias, "link a -> b", "rmlink a"} {
		_, ok := handler.Handle(reqenv, buildGuestReply(mention+text, nil)).(base.PlaceholderGuestResult)
		assert.False(t, ok, text)
	}
}

func TestGuestCommandsHandler_Help(t *testing.T) {
	handler := NewGuestCommandsHandler(test.BuildApplicationEnv(db))
	reqenv := test.BuildRequestEnv()

	result := handler.Handle(reqenv, buildGuestMessage(test.UID, mention+"помощь"))
	assert.Equal(t, reqenv.Lang.Tr(GuestHelpTr), answerMessage(t, result))
}

func buildGuestReply(text string, replyTo *tgbotapi.Message) *tgbotapi.Message {
	msg := buildGuestMessage(test.UID, text)
	msg.ReplyToMessage = replyTo
	return msg
}

// unwrapGuestResult returns the result that is sent to Telegram.
func unwrapGuestResult(result tgbotapi.InlineQueryResult) tgbotapi.InlineQueryResult {
	if wrapper, ok := result.(base.PlaceholderGuestResult); ok {
		return wrapper.Result
	}
	return result
}

func answerMessage(t *testing.T, result tgbotapi.InlineQueryResult) string {
	article, ok := unwrapGuestResult(result).(tgbotapi.InlineQueryResultArticle)
	assert.True(t, ok)
	content, ok := article.InputMessageContent.(tgbotapi.InputTextMessageContent)
	assert.True(t, ok)
	return content.Text
}

func stickerMessage() *tgbotapi.Message {
	return &tgbotapi.Message{Sticker: &tgbotapi.Sticker{FileID: test.FileID, FileUniqueID: test.UniqueFileID}}
}
