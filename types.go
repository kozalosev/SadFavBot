package main

import (
	"context"
	"database/sql"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
)

type appParams struct {
	ctx             context.Context
	messageHandlers []base.MessageHandler
	inlineHandlers  []base.InlineHandler
	api             *base.BotAPI
	stateStorage    wizard.StateStorage
	db              *sql.DB
}
