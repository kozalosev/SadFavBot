package common

import (
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/stretchr/testify/assert"
)

func TestReplyPossiblySelfDestroying_Ephemeral(t *testing.T) {
	bot := &base.FakeBotAPI{}
	appenv := &base.ApplicationEnv{Bot: bot}
	msg := &tgbotapi.Message{
		Chat:               tgbotapi.Chat{ID: -100, Type: "supergroup"},
		From:               &tgbotapi.User{ID: 1},
		EphemeralMessageID: 42,
	}

	// the reply goes through the usual way, which knows how to answer to ephemeral messages, without the deletion timer
	ReplyPossiblySelfDestroying(appenv, msg, "text", base.NoOpCustomizer)
	assert.Equal(t, []string{"text"}, bot.GetOutput())
}

func TestReplyToOriginal(t *testing.T) {
	plain := &tgbotapi.Message{MessageID: 10}
	assert.Same(t, plain, ReplyToOriginal(plain))

	reply := &tgbotapi.Message{MessageID: 10, ReplyToMessage: &tgbotapi.Message{MessageID: 3}}
	original := ReplyToOriginal(reply)
	assert.Equal(t, 3, original.MessageID)
	assert.Equal(t, 10, reply.MessageID, "the source message must stay intact")

	ephemeral := &tgbotapi.Message{ReplyToMessage: &tgbotapi.Message{MessageID: 3}, EphemeralMessageID: 42}
	assert.Same(t, ephemeral, ReplyToOriginal(ephemeral))
}
