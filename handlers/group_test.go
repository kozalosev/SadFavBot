package handlers

import (
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreUnknownInGroupsHandler_CanHandle(t *testing.T) {
	handler := NewIgnoreUnknownInGroupsHandler("TestMockBotAPI")
	reqenv := test.BuildRequestEnv()

	groupMsg := buildGroupCommand(test.UID, "text")
	assert.True(t, handler.CanHandle(reqenv, groupMsg))

	groupMsg.ReplyToMessage = &tgbotapi.Message{From: &tgbotapi.User{UserName: "someone"}}
	assert.True(t, handler.CanHandle(reqenv, groupMsg), "a reply to another user")

	groupMsg.ReplyToMessage = &tgbotapi.Message{From: &tgbotapi.User{UserName: "testmockbotapi"}}
	assert.False(t, handler.CanHandle(reqenv, groupMsg), "a reply to the bot may be an answer to a wizard")

	assert.False(t, handler.CanHandle(reqenv, buildMessage(test.UID)), "private chats")
}

func TestWizardsAskInGroups(t *testing.T) {
	registerWizards()
	appenv := test.BuildApplicationEnv(db)
	stateStorage := &wizard.FakeStorage{}

	tests := []struct {
		command string
		handler base.MessageHandler
		prompt  string
	}{
		{"/save", NewSaveHandler(appenv, stateStorage), SaveFieldsTrPrefix + FieldAlias},
		{"/link", NewLinkHandler(appenv, stateStorage), LinkFieldTrPrefix + FieldName},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			bot := appenv.Bot.(*base.FakeBotAPI)
			bot.ClearOutput()

			tt.handler.Handle(test.BuildRequestEnv(), buildGroupCommand(test.UID, tt.command))

			// the self-destruction is disabled in tests, so the question is sent by Reply
			assert.Equal(t, []string{tt.prompt}, bot.GetOutput())
		})
	}
}

func TestEphemeralCommandsHintInGroups(t *testing.T) {
	registerWizards()
	appenv := test.BuildApplicationEnv(db)
	stateStorage := &wizard.FakeStorage{}
	reqenv := buildRequestEnvWithStrings(map[string]string{
		DeleteGroupHintTr: "delete hint",
		ListGroupHintTr:   "list hint",
		RefGroupHintTr:    "ref hint",
		RmLinkGroupHintTr: "rmlink hint",
	})

	tests := []struct {
		command string
		handler base.MessageHandler
		hint    string
	}{
		{"/delete", NewDeleteHandler(appenv, stateStorage), "delete hint"},
		{"/list", NewListHandler(appenv, stateStorage), "list hint"},
		{"/ref", NewRefHandler(appenv, stateStorage), "ref hint"},
		{"/rmlink", NewRemoveLinkHandler(appenv, stateStorage), "rmlink hint"},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			bot := appenv.Bot.(*base.FakeBotAPI)
			bot.ClearOutput()
			msg := buildGroupCommand(test.UID, tt.command)
			msg.EphemeralMessageID = msg.MessageID // the commands are ephemeral in groups

			tt.handler.Handle(reqenv, msg)

			// an ephemeral reply disappears by itself, so it's sent as is, by Reply
			assert.Equal(t, []string{tt.hint}, bot.GetOutput())
		})
	}
}
