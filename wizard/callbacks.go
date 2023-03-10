package wizard

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	callbackDataSep         = ":"
	callbackDataFieldPrefix = "field" + callbackDataSep
	callbackDataErrorTr     = "callbacks.error"
	callbackDataSuccessTr   = "callbacks.was.set"
)

func CallbackQueryHandler(reqenv *base.RequestEnv, query *tgbotapi.CallbackQuery, stateStorage StateStorage) {
	id := query.From.ID
	var (
		form       Form
		err        error
		fieldValue string
	)
	if err = stateStorage.GetCurrentState(id, &form); err == nil {
		data := strings.TrimPrefix(query.Data, callbackDataFieldPrefix)
		dataArr := strings.Split(data, callbackDataSep)
		dataArrLen := len(dataArr)
		if dataArrLen == 2 {
			fieldName := dataArr[0]
			fieldValue = dataArr[1]
			field := form.Fields.FindField(fieldName)
			field.Data = fieldValue
			err = stateStorage.SaveState(id, &form)
		} else {
			err = errors.New(fmt.Sprintf("CallbackQuery data has %d fields unexpectedly!", dataArrLen))
		}
	}
	var c tgbotapi.Chattable
	if err != nil {
		c = tgbotapi.NewCallbackWithAlert(query.ID, reqenv.Lang.Tr(callbackDataErrorTr))
		if err = reqenv.Bot.Request(c); err != nil {
			log.Error(err)
		}
	} else {
		c = tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, reqenv.Lang.Tr(callbackDataSuccessTr)+fieldValue)

		msg := query.Message.ReplyToMessage
		form.PopulateRestored(msg, stateStorage)
		form.ProcessNextField(reqenv, msg)
	}
	if err := reqenv.Bot.Request(c); err != nil {
		log.Error(err)
	}
}
