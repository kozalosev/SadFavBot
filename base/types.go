/*
Package base defines basic types and a wrapper around the original [tgbotapi.BotAPI] struct.
*/
package base

import (
	"context"
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/loctools/go-l10n/loc"
)

// MessageHandler is a handler for the [tgbotapi.Message] update type.
type MessageHandler interface {
	CanHandle(msg *tgbotapi.Message) bool
	Handle(reqenv *RequestEnv, msg *tgbotapi.Message)
}

// InlineHandler is a handler for the [tgbotapi.InlineQuery] update type.
type InlineHandler interface {
	CanHandle(query *tgbotapi.InlineQuery) bool
	Handle(reqenv *RequestEnv, query *tgbotapi.InlineQuery)
}

// CallbackHandler is a handler for the [tgbotapi.CallbackQuery] update type.
type CallbackHandler interface {
	GetCallbackPrefix() string
	Handle(reqenv *RequestEnv, query *tgbotapi.CallbackQuery)
}

// BotAPI is a wrapper around the original [tgbotapi.BotAPI] struct.
type BotAPI struct {
	internal *tgbotapi.BotAPI

	// disables actual execution of requests
	// for testing purposes predominantly
	DummyMode bool
}

// RequestEnv is a container for all request related common resources. It's passed to all kinds of handlers.
type RequestEnv struct {
	// Bot is used when you need to send a request to Telegram Bot API.
	Bot *BotAPI
	// Lang is a localization container. You can get a message in the user's language by key, using its [loc.Context.Tr] method.
	Lang *loc.Context
	// Database is a reference to a [sql.DB] object.
	Database *sql.DB
	// Ctx is a context of the application; It's state will be switched to Done when the application is received the SIGTERM signal.
	Ctx context.Context
	// Options is a container for user options fetched from the database.
	Options *settings.UserOptions
}
