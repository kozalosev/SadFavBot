package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/goSadTgBot/base"
)

type IgnoreChatMemberJoinLeftHandler struct{}

func NewIgnoreChatMemberJoinLeftHandler() *IgnoreChatMemberJoinLeftHandler {
	return &IgnoreChatMemberJoinLeftHandler{}
}

func (*IgnoreChatMemberJoinLeftHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	return msg.NewChatMembers != nil || msg.LeftChatMember != nil
}

func (handler *IgnoreChatMemberJoinLeftHandler) Handle(_ *base.RequestEnv, _ *tgbotapi.Message) {}
