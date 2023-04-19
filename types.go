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

func newAppEnv(params *appParams) *base.ApplicationEnv {
	return &base.ApplicationEnv{
		Bot:      params.api,
		Database: params.db,
		Ctx:      params.ctx,
	}
}

func newRequestEnv(langCtx *loc.Context, opts *settings.UserOptions) *base.RequestEnv {
	return &base.RequestEnv{
		Lang:    langCtx,
		Options: opts,
	}
}
