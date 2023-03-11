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

func CallbackQueryHandler(reqenv *base.RequestEnv, stateStorage StateStorage) {
	id := reqenv.CallbackQuery.From.ID
	var (
		form       Form
		err        error
		fieldValue string
	)
	if err = stateStorage.GetCurrentState(id, &form); err == nil {
		data := strings.TrimPrefix(reqenv.CallbackQuery.Data, callbackDataFieldPrefix)
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
		c = tgbotapi.NewCallbackWithAlert(reqenv.CallbackQuery.ID, reqenv.Lang.Tr(callbackDataErrorTr))
		if err = reqenv.Bot.Request(c); err != nil {
			log.Error(err)
		}
	} else {
		msg := reqenv.CallbackQuery.Message
		c = tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, reqenv.Lang.Tr(callbackDataSuccessTr)+fieldValue)

		reqenv.Message = msg.ReplyToMessage
		form.PopulateRestored(reqenv, stateStorage)
		form.ProcessNextField(reqenv)
	}
	if err := reqenv.Bot.Request(c); err != nil {
		log.Error(err)
	}
}
