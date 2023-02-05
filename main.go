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
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const DefaultMessageTr = "commands.default.message"

var locpool = loc.NewPool("en")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
		}
	}

	wg.Wait()
	shutdown(stateStorage, db)
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
		handlers.HelpHandler{},
		handlers.SaveHandler{StateStorage: stateStorage},
		handlers.DeleteHandler{StateStorage: stateStorage},
		handlers.CancelHandler{StateStorage: stateStorage},
	}
	inlineHandlers = []base.InlineHandler{
		handlers.GetFavoritesInlineHandler{},
	}
	return
}

func processMessage(appParams *appParams, msg *tgbotapi.Message) {
	lc := locpool.GetContext(msg.From.LanguageCode)
	reqenv := &base.RequestEnv{
		Bot:      appParams.api,
		Message:  msg,
		Lang:     lc,
		Database: appParams.db,
		Ctx:      appParams.ctx,
	}

	for _, handler := range appParams.messageHandlers {
		if handler.CanHandle(msg) {
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

	reqenv.Reply(reqenv.Lang.Tr(DefaultMessageTr))
}

func processInline(appParams *appParams, query *tgbotapi.InlineQuery) {
	lc := locpool.GetContext(query.From.LanguageCode)
	reqenv := &base.RequestEnv{
		Bot:         appParams.api,
		InlineQuery: query,
		Lang:        lc,
		Database:    appParams.db,
		Ctx:         appParams.ctx,
	}

	for _, handler := range appParams.inlineHandlers {
		if handler.CanHandle(query) {
			handler.Handle(reqenv)
			return
		}
	}
}

func shutdown(stateStorage wizard.StateStorage, db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Errorln(err)
	}
	if err := stateStorage.Close(); err != nil {
		log.Errorln(err)
	}
}
