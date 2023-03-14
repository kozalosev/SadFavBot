package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type webhookParams struct {
	host string
	port string
	path string
}

func getWebhookParamsFromEnv() webhookParams {
	return webhookParams{
		host: os.Getenv("WEBHOOK_HOST"),
		port: os.Getenv("WEBHOOK_PORT"),
		path: strings.TrimPrefix(os.Getenv("WEBHOOK_PATH"), "/"),
	}
}

func setAndListenForWebhook(bot *tgbotapi.BotAPI, appParams *appParams, wg *sync.WaitGroup) *http.Server {
	whParams := getWebhookParamsFromEnv()
	path := fmt.Sprintf("/%s/%s", whParams.path, bot.Token)
	wh, err := tgbotapi.NewWebhook(fmt.Sprintf("https://%s:%s%s", whParams.host, whParams.port, path))
	if err != nil {
		panic(err)
	}
	if _, err := bot.Request(wh); err != nil {
		panic(err)
	}
	srv := &http.Server{Addr: ":" + whParams.port}
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		upd, err := bot.HandleUpdate(r)
		if err != nil {
			log.Error(err)
		} else {
			handleUpdate(appParams, wg, upd)
		}
	})
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	return srv
}

func shutdownWebhookServer(server *http.Server) {
	ctx, c := context.WithTimeout(context.Background(), time.Minute)
	defer c()
	if err := server.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}
