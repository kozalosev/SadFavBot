package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers"
	"github.com/kozalosev/SadFavBot/handlers/help"
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
	// the application is listening for the SIGTERM signal to exit
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addHttpHandlerForMetrics()
	srv := startServer(os.Getenv("APP_PORT"))

	stateStorage, db := establishConnections(ctx)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		panic(err)
	}
	api := base.NewBotAPI(bot)
	debugMode := os.Getenv("DEBUG")
	bot.Debug = strings.ToLower(debugMode) == "true" || debugMode == "1"

	appenv := &base.ApplicationEnv{
		Bot:      api,
		Database: db,
		Ctx:      ctx,
	}

	messageHandlers, inlineHandlers, callbackHandlers := initHandlers(appenv, stateStorage)

	appParams := &appParams{
		ctx:              ctx,
		messageHandlers:  messageHandlers,
		inlineHandlers:   inlineHandlers,
		callbackHandlers: callbackHandlers,
		settings:         repo.NewUserService(appenv),
		api:              api,
		stateStorage:     stateStorage,
		db:               db,
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

		// Unfortunately, the loop won't exit immediately.
		// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/207
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

	wg.Wait() // wait until all executing goroutines finish
	shutdown(stateStorage, db)
}

func establishConnections(ctx context.Context) (stateStorage wizard.StateStorage, db *pgxpool.Pool) {
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
	db = storage.ConnectToDatabase(ctx, dbConfig)
	storage.RunMigrations(dbConfig, os.Getenv("MIGRATIONS_REPO"))
	return
}

func initHandlers(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) (messageHandlers []base.MessageHandler, inlineHandlers []base.InlineHandler, callbackHandlers []base.CallbackHandler) {
	help.InitMessages(handlers.MaxAliasLen, handlers.MaxPackageNameLen, handlers.ReservedSymbolsForMessage)

	languageHandler := handlers.NewLanguageHandler(appenv, stateStorage)
	installPackageHandler := handlers.NewInstallPackageHandler(appenv, stateStorage)
	startEmbeddedHandlers := handlers.StartEmbeddedHandlers{
		Language:       &languageHandler,
		InstallPackage: &installPackageHandler,
	}

	messageHandlers = []base.MessageHandler{
		handlers.NewSaveHandler(appenv, stateStorage),
		handlers.NewListHandler(appenv, stateStorage),
		handlers.NewDeleteHandler(appenv, stateStorage),
		installPackageHandler,
		handlers.NewStartHandler(appenv, stateStorage, startEmbeddedHandlers),
		help.NewCommandHandler(appenv),
		handlers.NewCancelHandler(appenv, stateStorage),
		languageHandler,
		handlers.NewLinkHandler(appenv, stateStorage),
		handlers.NewPackageHandler(appenv, stateStorage),
		handlers.NewSearchModeHandler(appenv, stateStorage),
	}
	inlineHandlers = []base.InlineHandler{
		handlers.NewGetFavoritesInlineHandler(appenv),
	}
	callbackHandlers = []base.CallbackHandler{
		help.NewCallbackHandler(appenv),
	}
	registerMessageHandlerCounters(messageHandlers...)
	registerInlineHandlerCounters(inlineHandlers...)
	return
}

func startServer(port string) *http.Server {
	srv := &http.Server{Addr: ":" + port}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	return srv
}

func stopListeningForIncomingRequests(srv *http.Server) {
	ctx, c := context.WithTimeout(context.Background(), time.Minute)
	defer c()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}

func shutdown(stateStorage wizard.StateStorage, dbPool *pgxpool.Pool) {
	dbPool.Close()
	if err := stateStorage.Close(); err != nil {
		log.Error(err)
	}
}
