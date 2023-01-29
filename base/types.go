package base

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/loctools/go-l10n/loc"
)

type MessageHandler interface {
	CanHandle(msg *tgbotapi.Message) bool
	Handle(request *RequestEnv)
}

type InlineHandler interface {
	CanHandle(query *tgbotapi.InlineQuery) bool
	Handle(reqenv *RequestEnv)
}

type BotAPI struct {
	*tgbotapi.BotAPI

	DummyMode bool // for testing purposes predominantly
}

type RequestEnv struct {
	Bot         *BotAPI
	Message     *tgbotapi.Message
	InlineQuery *tgbotapi.InlineQuery
	Lang        *loc.Context
	Database    *sql.DB
}
