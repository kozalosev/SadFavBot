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

	addHttpHandlerForMetrics()
	srv := startServer(os.Getenv("APP_PORT"))

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

	var (
		wg         sync.WaitGroup
		wasStopped bool
	)
	if bot.Debug {
		if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
			panic(err)
		}

		updateConfig := tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}
		updates := bot.GetUpdatesChan(updateConfig)

		for upd := range updates {
			select {
			case <-ctx.Done():
				if !wasStopped {
					bot.StopReceivingUpdates()
					wasStopped = true
				}
			default:
			}
			handleUpdate(appParams, &wg, &upd)
		}
	} else {
		addHttpHandlerForWebhook(bot, appParams, &wg)
		<-ctx.Done()
		stopListeningForIncomingRequests(srv)
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
	dbConfig := storage.NewDatabaseConfig(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
	db = storage.ConnectToDatabase(dbConfig)
	storage.RunMigrations(dbConfig, os.Getenv("MIGRATIONS_REPO"))
	return
}

func initHandlers(stateStorage wizard.StateStorage) (messageHandlers []base.MessageHandler, inlineHandlers []base.InlineHandler) {
	installPackageHandler := handlers.InstallPackageHandler{StateStorage: stateStorage}
	messageHandlers = []base.MessageHandler{
		handlers.SaveHandler{StateStorage: stateStorage},
		handlers.ListHandler{StateStorage: stateStorage},
		handlers.DeleteHandler{StateStorage: stateStorage},
		installPackageHandler,
		handlers.StartHandler{StateStorage: stateStorage, InstallPackageHandler: &installPackageHandler},
		handlers.HelpHandler{},
		handlers.CancelHandler{StateStorage: stateStorage},
		handlers.LanguageHandler{StateStorage: stateStorage},
		handlers.LinkHandler{StateStorage: stateStorage},
		handlers.PackageHandler{StateStorage: stateStorage},
	}
	inlineHandlers = []base.InlineHandler{
		handlers.GetFavoritesInlineHandler{},
	}
	registerMessageHandlerCounters(messageHandlers...)
	registerInlineHandlerCounters(inlineHandlers...)
	return
}

func startServer(port string) *http.Server {
	srv := &http.Server{Addr: ":" + port}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()
	return srv
}

func stopListeningForIncomingRequests(srv *http.Server) {
	ctx, c := context.WithTimeout(context.Background(), time.Minute)
	defer c()
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorln(err)
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
