package wizard

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/logconst"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	// CallbackDataFieldPrefix is used in routing of callback updates.
	CallbackDataFieldPrefix = "field" + callbackDataSep

	callbackDataSep     = ":"
	callbackDataErrorTr = "callbacks.error"
)

// CallbackQueryHandler is a handler for callback updates generated by messages for fields with inline buttons.
func CallbackQueryHandler(reqenv *base.RequestEnv, query *tgbotapi.CallbackQuery, resources *Env) {
	id := query.From.ID
	var (
		form       Form
		err        error
		fieldValue string
	)
	if err = resources.stateStorage.GetCurrentState(id, &form); err == nil {
		data := strings.TrimPrefix(query.Data, CallbackDataFieldPrefix)
		dataArr := strings.Split(data, callbackDataSep)
		dataArrLen := len(dataArr)
		if dataArrLen == 2 {
			fieldName := dataArr[0]
			fieldValue = dataArr[1]
			field := form.Fields.FindField(fieldName)
			field.Data = fieldValue
			err = resources.stateStorage.SaveState(id, &form)
		} else {
			err = errors.New(fmt.Sprintf("CallbackQuery data has %d fields unexpectedly!", dataArrLen))
		}
	}
	var c tgbotapi.Chattable
	if err != nil {
		c = tgbotapi.NewCallbackWithAlert(query.ID, reqenv.Lang.Tr(callbackDataErrorTr))
		if err = resources.appEnv.Bot.Request(c); err != nil {
			log.WithField(logconst.FieldHandler, "wizard.CallbackQueryHandler").
				WithField(logconst.FieldCalledObject, "BotAPI").
				WithField(logconst.FieldCalledMethod, "Request").
				Error(err)
		}
	} else {
		chosenValue := reqenv.Lang.Tr(fieldValue)
		c = tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, query.Message.Text+" "+chosenValue)

		msg := query.Message.ReplyToMessage
		form.PopulateRestored(msg, resources)
		form.ProcessNextField(reqenv, msg)
	}
	if err := resources.appEnv.Bot.Request(c); err != nil {
		log.WithField(logconst.FieldHandler, "wizard.CallbackQueryHandler").
			WithField(logconst.FieldCalledObject, "BotAPI").
			WithField(logconst.FieldCalledMethod, "Request").
			Error(err)
	}
}
