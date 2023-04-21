package main

import (
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/logconst"
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
		}(*upd.InlineQuery) // copy by value
	} else if upd.ChosenInlineResult != nil {
		inc(chosenInlineResultCounter)
	} else if upd.Message != nil {
		wg.Add(1)
		go func(msg tgbotapi.Message) {
			defer wg.Done()
			processMessage(appParams, &msg)
		}(*upd.Message) // copy by value
	} else if upd.CallbackQuery != nil {
		wg.Add(1)
		go func(query tgbotapi.CallbackQuery) {
			defer wg.Done()
			processCallbackQuery(appParams, &query)
		}(*upd.CallbackQuery) // copy by value
	}
}

func processMessage(appParams *appParams, msg *tgbotapi.Message) {
	lang, opts := appParams.settings.FetchUserOptions(msg.From.ID, msg.From.LanguageCode)
	lc := locpool.GetContext(string(lang))
	reqenv := newRequestEnv(lc, opts)
	appenv := newAppEnv(appParams)

	// for commands and other handlers
	for _, handler := range appParams.messageHandlers {
		if handler.CanHandle(msg) {
			incMessageHandlerCounter(handler)
			handler.Handle(reqenv, msg)
			return
		}
	}

	// If no handler was chosen, check if this is a parameter for some previously created form.
	var form wizard.Form
	err := appParams.stateStorage.GetCurrentState(msg.From.ID, &form)
	if err == nil {
		resources := wizard.NewEnv(appenv, appParams.stateStorage)
		form.PopulateRestored(msg, resources)
		form.ProcessNextField(reqenv, msg)
		return
	}
	if err != redis.Nil {
		log.WithField(logconst.FieldFunc, "processMessage").
			Error("error occurred while getting current state: ", err)
		return
	}

	// fallback/default handler
	var defaultMessageTr string
	if msg.IsCommand() {
		defaultMessageTr = DefaultMessageOnCommandTr
	} else {
		defaultMessageTr = DefaultMessageTr
	}
	appenv.Bot.Reply(msg, reqenv.Lang.Tr(defaultMessageTr))
}

func processInline(appParams *appParams, query *tgbotapi.InlineQuery) {
	lang, opts := appParams.settings.FetchUserOptions(query.From.ID, query.From.LanguageCode)
	lc := locpool.GetContext(string(lang))
	reqenv := newRequestEnv(lc, opts)

	for _, handler := range appParams.inlineHandlers {
		if handler.CanHandle(query) {
			incInlineHandlerCounter(handler)
			handler.Handle(reqenv, query)
			return
		}
	}
}

func processCallbackQuery(appParams *appParams, query *tgbotapi.CallbackQuery) {
	lang, opts := appParams.settings.FetchUserOptions(query.From.ID, query.From.LanguageCode)
	lc := locpool.GetContext(string(lang))
	reqenv := newRequestEnv(lc, opts)

	splitData := strings.SplitN(query.Data, ":", 2)
	if len(splitData) < 2 {
		log.WithField(logconst.FieldFunc, "processCallbackQuery").
			Warningf("Unexpected callback: %+v", query)
		return
	}
	prefix := splitData[0] + ":"

	// special case for the wizard callback, otherwise check other [base.CallbackHandler]s
	if prefix == wizard.CallbackDataFieldPrefix {
		resources := wizard.NewEnv(newAppEnv(appParams), appParams.stateStorage)
		wizard.CallbackQueryHandler(reqenv, query, resources)
	} else {
		for _, handler := range appParams.callbackHandlers {
			if prefix == handler.GetCallbackPrefix() {
				handler.Handle(reqenv, query)
				return
			}
		}
	}
}
