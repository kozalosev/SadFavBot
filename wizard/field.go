package wizard

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"golang.org/x/exp/slices"
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
		reqenv.ReplyWithKeyboard(promptDescription, f.descriptor.InlineKeyboardAnswers)
	} else {
		reqenv.Reply(promptDescription)
	}
}

func (f *Field) validate(msg *tgbotapi.Message, lc *loc.Context) error {
	if len(f.descriptor.InlineKeyboardAnswers) > 0 && !slices.Contains(f.descriptor.InlineKeyboardAnswers, msg.Text) {
		return errors.New(ValidErrNotInListTr)
	}
	if f.descriptor.Validator == nil {
		return nil
	}
	return f.descriptor.Validator(msg, lc)
}
