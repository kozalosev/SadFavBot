package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"sync"
)

type webhookParams struct {
	host    string
	port    string
	path    string
	appPath string
}

func getWebhookParamsFromEnv() webhookParams {
	return webhookParams{
		host:    os.Getenv("WEBHOOK_HOST"),
		port:    os.Getenv("WEBHOOK_PORT"),
		path:    strings.TrimPrefix(os.Getenv("WEBHOOK_PATH"), "/"),
		appPath: strings.Trim(os.Getenv("APP_PATH"), "/"),
	}
}

func addHttpHandlerForWebhook(bot *tgbotapi.BotAPI, appParams *appParams, wg *sync.WaitGroup) {
	whParams := getWebhookParamsFromEnv()
	path := fmt.Sprintf("/%s/%s", whParams.path, bot.Token)
	whURL := fmt.Sprintf("https://%s:%s/%s%s", whParams.host, whParams.port, whParams.appPath, path)
	log.Info("Webhook URL: ", whURL[:len(bot.Token)], "/***")
	wh, err := tgbotapi.NewWebhook(whURL)
	if err != nil {
		panic(err)
	}
	if _, err := bot.Request(wh); err != nil {
		panic(err)
	}
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		upd, err := bot.HandleUpdate(r)
		if err != nil {
			log.Error(err)
		} else {
			handleUpdate(appParams, wg, upd)
		}
	})
}
