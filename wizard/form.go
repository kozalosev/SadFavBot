package wizard

import (
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
		descriptor: fieldDesc,
	}
	form.Fields = append(form.Fields, field)
}

func (form *Form) AddPrefilledField(name string, value interface{}) {
	field := &Field{Name: name, Data: value}
	form.Fields = append(form.Fields, field)
}

func (form *Form) ProcessNextField(reqenv *base.RequestEnv) {
	maxIndex := len(form.Fields) - 1
start:
	if form.Index > maxIndex {
		form.DoAction(reqenv)
		return
	}

	if form.Fields[form.Index].Data != nil || shouldBeSkipped(form.Fields[form.Index], form) {
		form.Index++
		goto start
	}

	currentField := form.Fields[form.Index]
	if currentField.WasRequested {
		value := currentField.extractor(reqenv.Message)
		if value == nil {
			reqenv.Reply(reqenv.Lang.Tr(InvalidFieldValueTypeErrorTr) + reqenv.Lang.Tr(string(currentField.Type)))
			return
		} else if err := currentField.validate(reqenv.Message, reqenv.Lang); err != nil {
			reqenv.Reply(reqenv.Lang.Tr(InvalidFieldValueErrorTr) + err.Error())
			return
		}
		currentField.Data = value
		form.Index++
		goto start
	} else {
		currentField.askUser(reqenv)
		currentField.WasRequested = true
	}

	err := form.storage.SaveState(reqenv.Message.From.ID, form)
	if err != nil {
		log.Error(err)
	}
}

func (form *Form) DoAction(reqenv *base.RequestEnv) {
	if form.descriptor.action == nil {
		reqenv.Reply(reqenv.Lang.Tr(MissingStateErrorTr))
		return
	}
	form.descriptor.action(reqenv, form.Fields)
}

func (form *Form) PopulateRestored(reqenv *base.RequestEnv, storage StateStorage) {
	form.storage = storage
	form.Fields[form.Index].restoreExtractor(reqenv.Message)
	form.descriptor = findFormDescriptor(form.WizardType)
	for _, field := range form.Fields {
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
