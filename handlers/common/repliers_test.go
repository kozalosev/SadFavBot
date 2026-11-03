package common

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/selfdestruct"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLang = loc.NewPool("en").GetContext("en")

func TestReplyPossiblySelfDestroying_Ephemeral(t *testing.T) {
	bot := &base.FakeBotAPI{}
	appenv := &base.ApplicationEnv{Bot: bot}
	msg := &tgbotapi.Message{
		Chat:               tgbotapi.Chat{ID: -100, Type: "supergroup"},
		From:               &tgbotapi.User{ID: 1},
		EphemeralMessageID: 42,
	}

	// the reply goes through the usual way, which knows how to answer to ephemeral messages, without the deletion timer
	ReplyPossiblySelfDestroying(appenv, testLang, msg, "text", base.NoOpCustomizer)
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

// startGroupReplyTest makes the self-destroying replies fast and returns the requests deleting the messages.
func startGroupReplyTest(t *testing.T, deletionAllowed bool) (*base.ApplicationEnv, *base.FakeBotAPI, func() []int) {
	bot := &base.FakeBotAPI{DeletionAllowed: deletionAllowed}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	selfdestruct.Start(ctx, bot, selfdestruct.NewMemoryQueue(), time.Millisecond)

	delay, warning := selfDeletionDelay, selfDeletionWarning
	selfDeletionDelay, selfDeletionWarning = time.Millisecond, time.Millisecond
	t.Cleanup(func() { selfDeletionDelay, selfDeletionWarning = delay, warning })

	deleted := func() []int {
		requests, _ := bot.GetOutput().([]tgbotapi.Chattable)
		var ids []int
		for _, req := range requests {
			switch deletion := req.(type) {
			case tgbotapi.DeleteMessageConfig:
				ids = append(ids, deletion.MessageID)
			case tgbotapi.DeleteMessagesConfig:
				ids = append(ids, deletion.MessageIDs...)
			}
		}
		return ids
	}
	return &base.ApplicationEnv{Bot: bot, Ctx: ctx}, bot, deleted
}

// The fake bot returns 0 as the identifier of every sent message, so the reply is the message 0.
func newGroupCommand() *tgbotapi.Message {
	return &tgbotapi.Message{
		MessageID:      10,
		Chat:           tgbotapi.Chat{ID: -100, Type: "supergroup"},
		From:           &tgbotapi.User{ID: 1},
		ReplyToMessage: &tgbotapi.Message{MessageID: 3},
	}
}

func TestReplyPossiblySelfDestroying_DeletesCommandIfAllowed(t *testing.T) {
	appenv, bot, deleted := startGroupReplyTest(t, true)
	ReplyPossiblySelfDestroying(appenv, testLang, newGroupCommand(), "text", base.NoOpCustomizer, ToOriginal)

	assert.Eventually(t, func() bool { return len(deleted()) == 2 }, time.Second, time.Millisecond)
	assert.ElementsMatch(t, []int{0, 10}, deleted(), "the reply and the command, not the original message")
	requests, _ := bot.GetOutput().([]tgbotapi.Chattable)
	reply, ok := requests[0].(tgbotapi.MessageConfig)
	require.True(t, ok, "%T", requests[0])
	assert.Equal(t, 3, reply.ReplyParameters.MessageID, "the reply goes to the original message")
	warning, ok := requests[1].(tgbotapi.EditMessageTextConfig)
	require.True(t, ok, "the reply is replaced with the warning before the deletion: %T", requests[1])
	assert.Equal(t, 0, warning.MessageID)
	assert.Equal(t, tgbotapi.ModeHTML, warning.ParseMode)
}

func TestReplyPossiblySelfDestroying_KeepsAllWithoutRights(t *testing.T) {
	appenv, _, deleted := startGroupReplyTest(t, false)
	ReplyPossiblySelfDestroying(appenv, testLang, newGroupCommand(), "text", base.NoOpCustomizer)

	assert.Never(t, func() bool { return len(deleted()) > 0 }, 30*time.Millisecond, time.Millisecond)
}
