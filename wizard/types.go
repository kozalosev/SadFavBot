package wizard

import (
	"github.com/kozalosev/SadFavBot/base"
)

type FormAction func(request *base.RequestEnv, fields Fields)

type Wizard interface {
	AddEmptyField(name string, fieldType FieldType)
	AddPrefilledField(name string, value interface{})
	ProcessNextField(reqenv *base.RequestEnv)
	DoAction(reqenv *base.RequestEnv)
	PopulateRestored(reqenv *base.RequestEnv, storage StateStorage)
}

//goland:noinspection GoNameStartsWithPackageName
type WizardMessageHandler interface {
	base.MessageHandler

	GetWizardName() string
	GetWizardStateStorage() StateStorage
	GetWizardDescriptor() *FormDescriptor
}
