package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
)

type IgnoreUnknownInGroupsHandler struct{}

func NewIgnoreUnknownInGroupsHandler() *IgnoreUnknownInGroupsHandler {
	return &IgnoreUnknownInGroupsHandler{}
}

func (*IgnoreUnknownInGroupsHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	return common.IsGroup(msg.Chat)
}

func (handler *IgnoreUnknownInGroupsHandler) Handle(_ *base.RequestEnv, _ *tgbotapi.Message) {}
