package wizard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type FieldValidator func(msg *tgbotapi.Message) error

type Fields []*Field
type FieldType string

const (
	Auto      FieldType = "<auto>"
	Text                = "text"
	Sticker             = "sticker"
	Image               = "image"
	Voice               = "voice"
	Audio               = "audio"
	Video               = "video"
	VideoNote           = "video_note"
	Gif                 = "gif"
)

type Field struct {
	Name                  string
	Data                  interface{}
	WasRequested          bool
	Type                  FieldType
	PromptDescription     string
	InlineKeyboardAnswers []string
	SkipIf                *SkipConditionContainer

	validator FieldValidator
	extractor FieldExtractor
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
	if len(f.InlineKeyboardAnswers) > 0 {
		reqenv.ReplyWithKeyboard(f.PromptDescription, f.InlineKeyboardAnswers)
	} else {
		reqenv.Reply(f.PromptDescription)
	}
}

func (f *Field) validate(msg *tgbotapi.Message) error {
	if f.validator == nil {
		return nil
	}
	return f.validator(msg)
}
