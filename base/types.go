// Package base defines basic types and a wrapper around the original [tgbotapi.BotAPI] struct.
package base

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

// ExtendedBotAPI is a set of more convenient methods to work with Telegram Bot API.
type ExtendedBotAPI interface {
	// GetName returns the name of the bot got from the getMe() request.
	GetName() string
	// ReplyWithMessageCustomizer is the most common method to send text messages as a reply. Use this method if you want
	// to change several options like a message in Markdown with an inline keyboard.
	ReplyWithMessageCustomizer(*tgbotapi.Message, string, MessageCustomizer)
	// Reply with just a text message, without any customizations.
	Reply(msg *tgbotapi.Message, text string)
	ReplyWithMarkdown(msg *tgbotapi.Message, text string)
	// ReplyWithKeyboard uses a one time reply keyboard.
	// https://core.telegram.org/bots/api#replykeyboardmarkup
	ReplyWithKeyboard(msg *tgbotapi.Message, text string, options []string)
	// ReplyWithInlineKeyboard attaches an inline keyboard to the message.
	// https://core.telegram.org/bots/api#inlinekeyboardmarkup
	ReplyWithInlineKeyboard(msg *tgbotapi.Message, text string, buttons []tgbotapi.InlineKeyboardButton)
	// Request is the most common method that can be used to send any request to Telegram.
	Request(tgbotapi.Chattable) error
}

// BotAPI is a wrapper around the original [tgbotapi.BotAPI] struct.
// It implements the [ExtendedBotAPI] interface.
type BotAPI struct {
	internal *tgbotapi.BotAPI
}

// ApplicationEnv is a container for all application scoped resources.
type ApplicationEnv struct {
	// Bot is used when you need to send a request to Telegram Bot API.
	Bot ExtendedBotAPI
	// Database is a reference to a [sql.DB]-like object.
	Database *pgxpool.Pool
	// Ctx is a context of the application; It's state will be switched to Done when the application is received the SIGTERM signal.
	Ctx context.Context
}

// RequestEnv is a container for all request related common resources. It's passed to all kinds of handlers.
type RequestEnv struct {
	// Lang is a localization container. You can get a message in the user's language by key, using its [loc.Context.Tr] method.
	Lang *loc.Context
	// Options is a container for user options fetched from the database.
	Options *settings.UserOptions
}
