package handlers

import (
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

// IgnoreUnknownInGroupsHandler swallows the messages in groups no other handler has taken, except replies to the bot:
// they may be answers to the questions of a wizard, which the framework looks for after all handlers.
type IgnoreUnknownInGroupsHandler struct {
	botName string
}

func NewIgnoreUnknownInGroupsHandler(botName string) *IgnoreUnknownInGroupsHandler {
	return &IgnoreUnknownInGroupsHandler{botName: botName}
}

func (handler *IgnoreUnknownInGroupsHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	return base.IsGroupChat(&msg.Chat) && !base.IsReplyToBot(msg, handler.botName)
}

func (handler *IgnoreUnknownInGroupsHandler) Handle(_ *base.RequestEnv, _ *tgbotapi.Message) {}

// IgnoreForeignCommandsHandler swallows commands addressed to other bots (/help@AnotherBot).
// The bot receives them when it's an admin of a group, but the standard command matching ignores the "@suffix".
type IgnoreForeignCommandsHandler struct {
	botName string
}

func NewIgnoreForeignCommandsHandler(botName string) *IgnoreForeignCommandsHandler {
	return &IgnoreForeignCommandsHandler{botName: botName}
}

func (handler *IgnoreForeignCommandsHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	_, target, found := strings.Cut(msg.CommandWithAt(), "@")
	return found && !strings.EqualFold(target, handler.botName)
}

func (*IgnoreForeignCommandsHandler) Handle(_ *base.RequestEnv, _ *tgbotapi.Message) {}
