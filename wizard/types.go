package wizard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
)

type FormAction func(request *base.RequestEnv, msg *tgbotapi.Message, fields Fields)

type Wizard interface {
	AddEmptyField(name string, fieldType FieldType)
	AddPrefilledField(name string, value interface{})
	ProcessNextField(reqenv *base.RequestEnv, msg *tgbotapi.Message)
	DoAction(reqenv *base.RequestEnv, msg *tgbotapi.Message)
	PopulateRestored(msg *tgbotapi.Message, storage StateStorage)
}

//goland:noinspection GoNameStartsWithPackageName
type WizardMessageHandler interface {
	base.MessageHandler

	GetWizardName() string
	GetWizardStateStorage() StateStorage
	GetWizardDescriptor() *FormDescriptor
}
