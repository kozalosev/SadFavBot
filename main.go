package main

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/handlers"
	"github.com/kozalosev/SadFavBot/storage"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultMessageTr          = "commands.default.message"
	DefaultMessageOnCommandTr = "commands.default.message.on.command"
)

var locpool = loc.NewPool("en")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metricsServer := startMetricsServer(os.Getenv("METRICS_PORT"))
	stateStorage, db := establishConnections(ctx)
	messageHandlers, inlineHandlers := initHandlers(stateStorage)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		panic(err)
	}
	api := base.NewBotAPI(bot)
	debugMode := os.Getenv("DEBUG")
	bot.Debug = strings.ToLower(debugMode) == "true" || debugMode == "1"

	appParams := &appParams{
		ctx:             ctx,
		messageHandlers: messageHandlers,
		inlineHandlers:  inlineHandlers,
		api:             api,
		stateStorage:    stateStorage,
		db:              db,
	}

	if wasPopulated := wizard.PopulateWizardDescriptors(messageHandlers); !wasPopulated {
		log.Warning("Wizard actions map already has been populated; skipping...")
	}

	updateConfig := tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}
	updates := bot.GetUpdatesChan(updateConfig)

	var (
		wg         sync.WaitGroup
		wasStopped bool
	)
	for upd := range updates {
		select {
		case <-ctx.Done():
			if !wasStopped {
				bot.StopReceivingUpdates()
				wasStopped = true
			}
		default:
		}
		if upd.InlineQuery != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				processInline(appParams, upd.InlineQuery)
			}()
		} else if upd.Message != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				processMessage(appParams, upd.Message)
			}()
		} else if upd.ChosenInlineResult != nil {
			inc(chosenInlineResultCounter)
		}
	}

	wg.Wait()
	shutdown(stateStorage, db, metricsServer)
}

func establishConnections(ctx context.Context) (stateStorage wizard.StateStorage, db *sql.DB) {
	commandStateTTL, err := time.ParseDuration(os.Getenv("COMMAND_STATE_TTL"))
	if err != nil {
		panic(err)
	}
	stateStorage = wizard.ConnectToRedis(ctx, commandStateTTL, &redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	db = storage.ConnectToDatabase(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
	return
}

func initHandlers(stateStorage wizard.StateStorage) (messageHandlers []base.MessageHandler, inlineHandlers []base.InlineHandler) {
	messageHandlers = []base.MessageHandler{
		handlers.SaveHandler{StateStorage: stateStorage},
		handlers.ListHandler{},
		handlers.DeleteHandler{StateStorage: stateStorage},
		handlers.StartHandler{StateStorage: stateStorage},
		handlers.HelpHandler{},
		handlers.CancelHandler{StateStorage: stateStorage},
		handlers.LanguageHandler{StateStorage: stateStorage},
	}
	inlineHandlers = []base.InlineHandler{
		handlers.GetFavoritesInlineHandler{},
	}
	registerMessageHandlerCounters(messageHandlers...)
	registerInlineHandlerCounters(inlineHandlers...)
	return
}

func processMessage(appParams *appParams, msg *tgbotapi.Message) {
	langCode := fetchLanguage(appParams.db, msg.From.ID, msg.From.LanguageCode)
	lc := locpool.GetContext(langCode)
	reqenv := &base.RequestEnv{
		Bot:      appParams.api,
		Message:  msg,
		Lang:     lc,
		Database: appParams.db,
		Ctx:      appParams.ctx,
	}

	for _, handler := range appParams.messageHandlers {
		if handler.CanHandle(msg) {
			incMessageHandlerCounter(handler)
			handler.Handle(reqenv)
			return
		}
	}

	var form wizard.Form
	err := appParams.stateStorage.GetCurrentState(msg.From.ID, &form)
	if err == nil {
		form.PopulateRestored(reqenv, appParams.stateStorage)
		form.ProcessNextField(reqenv)
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
	reqenv.Reply(reqenv.Lang.Tr(defaultMessageTr))
}

func processInline(appParams *appParams, query *tgbotapi.InlineQuery) {
	langCode := fetchLanguage(appParams.db, query.From.ID, query.From.LanguageCode)
	lc := locpool.GetContext(langCode)
	reqenv := &base.RequestEnv{
		Bot:         appParams.api,
		InlineQuery: query,
		Lang:        lc,
		Database:    appParams.db,
		Ctx:         appParams.ctx,
	}

	for _, handler := range appParams.inlineHandlers {
		if handler.CanHandle(query) {
			incInlineHandlerCounter(handler)
			handler.Handle(reqenv)
			return
		}
	}
}

func shutdown(stateStorage wizard.StateStorage, db *sql.DB, metricsServer *http.Server) {
	if err := db.Close(); err != nil {
		log.Errorln(err)
	}
	if err := stateStorage.Close(); err != nil {
		log.Errorln(err)
	}
	shutdownMetricsServer(metricsServer)
}
