package base

import (
	"context"
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/loctools/go-l10n/loc"
)

type MessageHandler interface {
	CanHandle(msg *tgbotapi.Message) bool
	Handle(reqenv *RequestEnv, msg *tgbotapi.Message)
}

type InlineHandler interface {
	CanHandle(query *tgbotapi.InlineQuery) bool
	Handle(reqenv *RequestEnv, query *tgbotapi.InlineQuery)
}

type CallbackHandler interface {
	GetCallbackPrefix() string
	Handle(reqenv *RequestEnv, query *tgbotapi.CallbackQuery)
}

type BotAPI struct {
	internal *tgbotapi.BotAPI

	DummyMode bool // for testing purposes predominantly
}

type RequestEnv struct {
	Bot      *BotAPI
	Lang     *loc.Context
	Database *sql.DB
	Ctx      context.Context
	Options  *settings.UserOptions
}
