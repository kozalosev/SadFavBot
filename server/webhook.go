package server

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/app"
	"github.com/kozalosev/SadFavBot/logconst"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"sync"
)

// AddHttpHandlerForWebhook uses [http.HandleFunc] to add a global route to the server.
func AddHttpHandlerForWebhook(bot *tgbotapi.BotAPI, appParams *app.Params, wg *sync.WaitGroup) {
	whParams := getWebhookParamsFromEnv()
	path := fmt.Sprintf("/%s/%s", whParams.path, bot.Token)
	whURL := fmt.Sprintf("https://%s:%s/%s%s", whParams.host, whParams.port, whParams.appPath, path)
	log.WithField(logconst.FieldFunc, "addHttpHandlerForWebhook").
		Info("Webhook URL: ", whURL[:len(bot.Token)], "/***")
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
			log.WithField(logconst.FieldFunc, "addHttpHandlerForWebhook").
				WithField(logconst.FieldCalledObject, "BotAPI").
				WithField(logconst.FieldCalledMethod, "HandleUpdate").
				Error(err)
		} else {
			app.HandleUpdate(appParams, wg, upd)
		}
	})
}

// Read comments in the `.env` file and
// https://github.com/kozalosev/SadFavBot/wiki/Run-and-configuration#on-a-server-production-mode
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
