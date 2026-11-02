package handlers

import (
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
)

type IgnoreUnknownInGroupsHandler struct{}

func NewIgnoreUnknownInGroupsHandler() *IgnoreUnknownInGroupsHandler {
	return &IgnoreUnknownInGroupsHandler{}
}

func (*IgnoreUnknownInGroupsHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	return common.IsGroup(&msg.Chat)
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
