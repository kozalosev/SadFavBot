package wizard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
)

const (
	InvalidFieldValueErrorTr     = "wizard.errors.field.invalid.value"
	InvalidFieldValueTypeErrorTr = "wizard.errors.field.invalid.type"
	MissingStateErrorTr          = "wizard.errors.state.missing"
)

type Form struct {
	Fields     Fields `json:"fields"`
	Index      int    `json:"index"`
	WizardType string `json:"wizardType"`

	storage    StateStorage
	descriptor *FormDescriptor
}

func (form *Form) AddEmptyField(name string, fieldType FieldType) {
	if form.descriptor == nil {
		panic("No descriptor was set for the form: " + form.WizardType)
	}
	fieldDesc := form.descriptor.findFieldDescriptor(name)
	if fieldDesc == nil {
		panic("No descriptor was set for the field: " + name)
	}
	field := &Field{
		Name:       name,
		Type:       fieldType,
		Form:       form,
		descriptor: fieldDesc,
	}
	form.Fields = append(form.Fields, field)
}

func (form *Form) AddPrefilledField(name string, value interface{}) {
	field := &Field{Name: name, Data: value, Form: form}
	form.Fields = append(form.Fields, field)
}

func (form *Form) ProcessNextField(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	maxIndex := len(form.Fields) - 1
start:
	if form.Index > maxIndex {
		form.DoAction(reqenv, msg)
		return
	}

	if form.Fields[form.Index].Data != nil || shouldBeSkipped(form.Fields[form.Index], form) {
		form.Index++
		goto start
	}

	currentField := form.Fields[form.Index]
	if currentField.WasRequested {
		value := currentField.extractor(msg)
		if value == nil {
			reqenv.Bot.Reply(msg, reqenv.Lang.Tr(InvalidFieldValueTypeErrorTr)+reqenv.Lang.Tr(string(currentField.Type)))
			return
		} else if err := currentField.validate(reqenv, msg); err != nil {
			reqenv.Bot.Reply(msg, reqenv.Lang.Tr(InvalidFieldValueErrorTr)+reqenv.Lang.Tr(err.Error()))
			return
		}
		currentField.Data = value
		form.Index++
		goto start
	} else {
		currentField.askUser(reqenv, msg)
		currentField.WasRequested = true
	}

	err := form.storage.SaveState(msg.From.ID, form)
	if err != nil {
		log.Error(err)
	}
}

func (form *Form) DoAction(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	if form.descriptor.action == nil {
		reqenv.Bot.Reply(msg, reqenv.Lang.Tr(MissingStateErrorTr))
		return
	}
	form.descriptor.action(reqenv, msg, form.Fields)
}

func (form *Form) PopulateRestored(msg *tgbotapi.Message, storage StateStorage) {
	form.storage = storage
	form.Fields[form.Index].restoreExtractor(msg)
	form.descriptor = findFormDescriptor(form.WizardType)
	for _, field := range form.Fields {
		field.Form = form
		field.descriptor = form.descriptor.findFieldDescriptor(field.Name)
	}
}

func NewWizard(handler WizardMessageHandler, fields int) Wizard {
	return &Form{
		storage:    handler.GetWizardStateStorage(),
		Fields:     make(Fields, 0, fields),
		WizardType: handler.GetWizardName(),
		descriptor: findFormDescriptor(handler.GetWizardName()),
	}
}

func shouldBeSkipped(field *Field, form *Form) bool {
	skipPredicate := field.descriptor.SkipIf
	if skipPredicate == nil {
		return false
	} else {
		return skipPredicate.ShouldBeSkipped(form)
	}
}
