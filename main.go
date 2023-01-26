package main

import (
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/handlers"
	"github.com/kozalosev/SadFavBot/storage"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

var messageHandlers = []base.MessageHandler{
	handlers.HelpHandler{},
	handlers.SaveHandler{StateStorage: stateStorage},
}
var inlineHandlers = []base.InlineHandler{
	handlers.InlineHandler{},
}

var locpool = loc.NewPool("en")
var stateStorage = wizard.RedisStateStorage{RDB: redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_ADDR"),
	Password: os.Getenv("REDIS_PASSWORD"),
	DB:       0,
})}
var DatabaseConn = storage.ConnectToDatabase(
	os.Getenv("POSTGRES_HOST"),
	os.Getenv("POSTGRES_USER"),
	os.Getenv("POSTGRES_PASSWORD"),
	os.Getenv("POSTGRES_DB"))

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		panic(err)
	}
	debugMode := os.Getenv("DEBUG")
	bot.Debug = strings.ToLower(debugMode) == "true" || debugMode == "1"

	wizard.PopulateWizardActions(map[string]wizard.FormAction{
		"SaveWizard": handlers.SaveFormToDB,
	})

	updateConfig := tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}
	updates := bot.GetUpdatesChan(updateConfig)

	for upd := range updates {
		if upd.InlineQuery == nil && upd.Message == nil {
			continue
		}

		api := &base.BotAPI{BotAPI: bot}

		if upd.InlineQuery != nil {
			processInline(api, upd.InlineQuery)
		} else if upd.Message != nil {
			processMessage(api, upd.Message)
		}
	}
}

func processMessage(api *base.BotAPI, msg *tgbotapi.Message) {
	lc := locpool.GetContext(msg.From.LanguageCode)
	reqenv := &base.RequestEnv{
		Bot:      api,
		Message:  msg,
		Lang:     lc,
		Database: DatabaseConn,
	}

	for _, handler := range messageHandlers {
		if handler.CanHandle(msg) {
			go handler.Handle(reqenv)
			return
		}
	}

	var form wizard.Form
	err := stateStorage.GetCurrentState(msg.From.ID, &form)
	if err == nil {
		form.PopulateRestored(reqenv, stateStorage)
		form.ProcessNextField(reqenv)
	} else {
		log.Errorln("nil form was restored: ", err)
	}
}

func processInline(api *base.BotAPI, query *tgbotapi.InlineQuery) {
	lc := locpool.GetContext(query.From.LanguageCode)
	reqenv := &base.RequestEnv{
		Bot:         api,
		InlineQuery: query,
		Lang:        lc,
		Database:    DatabaseConn,
	}

	for _, handler := range inlineHandlers {
		if handler.CanHandle(query) {
			go handler.Handle(reqenv)
			return
		}
	}
}
