package app

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

// Params is a huge container will all possible resources of the application.
// It should be used in the main function, app and server packages only!
type Params struct {
	Ctx              context.Context
	MessageHandlers  []base.MessageHandler
	InlineHandlers   []base.InlineHandler
	CallbackHandlers []base.CallbackHandler
	Settings         settings.OptionsFetcher
	LangPool         *loc.Pool
	API              *base.BotAPI
	StateStorage     wizard.StateStorage
	DB               *pgxpool.Pool
}

// NewAppEnv is a constructor for [base.ApplicationEnv].
// They reside in different packages to eliminate an import cycle.
func NewAppEnv(params *Params) *base.ApplicationEnv {
	return &base.ApplicationEnv{
		Bot:      params.API,
		Database: params.DB,
		Ctx:      params.Ctx,
	}
}
