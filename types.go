package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

type appParams struct {
	ctx              context.Context
	messageHandlers  []base.MessageHandler
	inlineHandlers   []base.InlineHandler
	callbackHandlers []base.CallbackHandler
	settings         settings.OptionsFetcher
	api              *base.BotAPI
	stateStorage     wizard.StateStorage
	db               *pgxpool.Pool
}

func newRequestEnv(params *appParams, langCtx *loc.Context, opts *settings.UserOptions) *base.RequestEnv {
	return &base.RequestEnv{
		Bot:      params.api,
		Lang:     langCtx,
		Database: params.db,
		Ctx:      params.ctx,
		Options:  opts,
	}
}
