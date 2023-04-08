package main

import (
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

func handleUpdate(appParams *appParams, wg *sync.WaitGroup, upd *tgbotapi.Update) {
	if upd.InlineQuery != nil {
		wg.Add(1)
		go func(query tgbotapi.InlineQuery) {
			defer wg.Done()
			processInline(appParams, &query)
		}(*upd.InlineQuery)
	} else if upd.ChosenInlineResult != nil {
		inc(chosenInlineResultCounter)
	} else if upd.Message != nil {
		wg.Add(1)
		go func(msg tgbotapi.Message) {
			defer wg.Done()
			processMessage(appParams, &msg)
		}(*upd.Message)
	} else if upd.CallbackQuery != nil {
		wg.Add(1)
		go func(query tgbotapi.CallbackQuery) {
			defer wg.Done()
			processCallbackQuery(appParams, &query)
		}(*upd.CallbackQuery)
	}
}

func processMessage(appParams *appParams, msg *tgbotapi.Message) {
	lang, opts := settings.FetchUserOptions(appParams.ctx, appParams.db, msg.From.ID, msg.From.LanguageCode)
	lc := locpool.GetContext(string(lang))
	reqenv := newRequestEnv(appParams, lc, opts)

	for _, handler := range appParams.messageHandlers {
		if handler.CanHandle(msg) {
			incMessageHandlerCounter(handler)
			handler.Handle(reqenv, msg)
			return
		}
	}

	var form wizard.Form
	err := appParams.stateStorage.GetCurrentState(msg.From.ID, &form)
	if err == nil {
		form.PopulateRestored(msg, appParams.stateStorage)
		form.ProcessNextField(reqenv, msg)
		return
	}
	if err != redis.Nil {
		log.Errorln("error occurred while getting current state: ", err)
		return
	}

	var defaultMessageTr string
	if msg.IsCommand() {
		defaultMessageTr = DefaultMessageOnCommandTr
	} else {
		defaultMessageTr = DefaultMessageTr
	}
	reqenv.Bot.Reply(msg, reqenv.Lang.Tr(defaultMessageTr))
}

func processInline(appParams *appParams, query *tgbotapi.InlineQuery) {
	lang, opts := settings.FetchUserOptions(appParams.ctx, appParams.db, query.From.ID, query.From.LanguageCode)
	lc := locpool.GetContext(string(lang))
	reqenv := newRequestEnv(appParams, lc, opts)

	for _, handler := range appParams.inlineHandlers {
		if handler.CanHandle(query) {
			incInlineHandlerCounter(handler)
			handler.Handle(reqenv, query)
			return
		}
	}
}

func processCallbackQuery(appParams *appParams, query *tgbotapi.CallbackQuery) {
	lang, opts := settings.FetchUserOptions(appParams.ctx, appParams.db, query.From.ID, query.From.LanguageCode)
	lc := locpool.GetContext(string(lang))
	reqenv := newRequestEnv(appParams, lc, opts)

	splitData := strings.SplitN(query.Data, ":", 2)
	if len(splitData) < 2 {
		log.Warningf("Unexpected callback: %+v", query)
		return
	}
	prefix := splitData[0] + ":"

	if prefix == wizard.CallbackDataFieldPrefix {
		wizard.CallbackQueryHandler(reqenv, query, appParams.stateStorage)
	} else {
		for _, handler := range appParams.callbackHandlers {
			if prefix == handler.GetCallbackPrefix() {
				handler.Handle(reqenv, query)
				return
			}
		}
	}
}
