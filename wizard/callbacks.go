package wizard

import (
	"fmt"
	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type CallbackQueryHandler func(reqenv *base.RequestEnv, storage StateStorage)

type callbackHandlerStruct struct {
	handler CallbackQueryHandler
	addedAt time.Time
}

var (
	mtx              sync.Mutex
	callbackHandlers = make(map[string]callbackHandlerStruct)
)

func AddCallbackHandler(msg *tgbotapi.Message, handler CallbackQueryHandler) {
	addCallbackHandler(buildKey(msg), handler)
}

func addCallbackHandler(messageID string, handler CallbackQueryHandler) {
	if _, contains := callbackHandlers[messageID]; contains {
		log.Warningf("CallbackQuery handler for %s is already in the map!", messageID)
	}
	mtx.Lock()
	callbackHandlers[messageID] = callbackHandlerStruct{
		handler: handler,
		addedAt: time.Now(),
	}
	mtx.Unlock()
}

func DeleteCallbackHandler(msg *tgbotapi.Message) {
	deleteCallbackHandler(buildKey(msg))
}

func deleteCallbackHandler(messageID string) {
	mtx.Lock()
	delete(callbackHandlers, messageID)
	mtx.Unlock()
}

func GetCallbackHandler(msg *tgbotapi.Message) (CallbackQueryHandler, bool) {
	return getCallbackHandler(buildKey(msg))
}

func getCallbackHandler(messageID string) (CallbackQueryHandler, bool) {
	var handler CallbackQueryHandler
	hs, ok := callbackHandlers[messageID]
	if ok {
		handler = hs.handler
	}
	return handler, ok
}

func buildKey(msg *tgbotapi.Message) string {
	return fmt.Sprintf("%d:%d", msg.Chat.ID, msg.MessageID)
}

func init() {
	scheduler := gocron.NewScheduler(time.UTC)
	_, err := scheduler.Every(1).Hour().Do(func() {
		mtx.Lock()
		for k, v := range callbackHandlers {
			if v.addedAt.Before(time.Now().Add(-time.Hour)) {
				delete(callbackHandlers, k)
			}
		}
		mtx.Unlock()
	})
	if err != nil {
		panic(err)
	}
	scheduler.StartAsync()
}
