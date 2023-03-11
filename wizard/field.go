package wizard

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"golang.org/x/exp/slices"
	"strings"
)

const ValidErrNotInListTr = "errors.validation.option.not.in.list"

type FieldValidator func(msg *tgbotapi.Message, lc *loc.Context) error

type Fields []*Field
type FieldType string

const (
	Auto      FieldType = "<auto>"
	Text      FieldType = "text"
	Sticker   FieldType = "sticker"
	Image     FieldType = "image"
	Voice     FieldType = "voice"
	Audio     FieldType = "audio"
	Video     FieldType = "video"
	VideoNote FieldType = "video_note"
	Gif       FieldType = "gif"

	callbackDataSep         = ":"
	callbackDataFieldPrefix = "field" + callbackDataSep
)

type Field struct {
	Name         string      `json:"name"`
	Data         interface{} `json:"data,omitempty"`
	WasRequested bool        `json:"wasRequested"`
	Type         FieldType   `json:"type"`

	extractor  FieldExtractor
	descriptor *FieldDescriptor
}

func (fs Fields) FindField(name string) *Field {
	found := funk.Filter(fs, func(f *Field) bool { return f.Name == name }).([]*Field)
	if len(found) == 0 {
		return nil
	}
	if len(found) > 1 {
		log.Warning("More than needed: ", found)
	}
	return found[0]
}

func (f *Field) askUser(reqenv *base.RequestEnv) {
	promptDescription := reqenv.Lang.Tr(f.descriptor.promptDescription)
	if len(f.descriptor.InlineKeyboardAnswers) > 0 {
		inlineAnswers := funk.Map(f.descriptor.InlineKeyboardAnswers, func(s string) base.InlineButton {
			return base.InlineButton{
				Text: s,
				Data: callbackDataFieldPrefix + f.Name + callbackDataSep + s,
			}
		}).([]base.InlineButton)
		sendMessage := reqenv.ReplyWithInlineKeyboard(promptDescription, inlineAnswers)
		if sendMessage != nil {
			AddCallbackHandler(sendMessage, callbackQueryHandler)
		}
	} else if len(f.descriptor.ReplyKeyboardAnswers) > 0 {
		reqenv.ReplyWithKeyboard(promptDescription, f.descriptor.ReplyKeyboardAnswers)
	} else {
		reqenv.Reply(promptDescription)
	}
}

func callbackQueryHandler(reqenv *base.RequestEnv, stateStorage StateStorage) {
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
		c = tgbotapi.NewCallbackWithAlert(reqenv.CallbackQuery.ID, reqenv.Lang.Tr("callbacks.error"))
		if err = reqenv.Bot.Request(c); err != nil {
			log.Error(err)
		}
	} else {
		msg := reqenv.CallbackQuery.Message
		c = tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, reqenv.Lang.Tr("callbacks.was.set")+fieldValue)

		reqenv.Message = msg.ReplyToMessage
		form.PopulateRestored(reqenv, stateStorage)
		form.ProcessNextField(reqenv)
	}
	if err := reqenv.Bot.Request(c); err != nil {
		log.Error(err)
	}
}

func (f *Field) validate(msg *tgbotapi.Message, lc *loc.Context) error {
	if len(f.descriptor.ReplyKeyboardAnswers) > 0 && !slices.Contains(f.descriptor.ReplyKeyboardAnswers, msg.Text) {
		return errors.New(ValidErrNotInListTr)
	}
	if f.descriptor.Validator == nil {
		return nil
	}
	return f.descriptor.Validator(msg, lc)
}
