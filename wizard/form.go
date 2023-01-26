package wizard

import (
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
)

type FormAction func(request *base.RequestEnv, fields Fields)

type Wizard interface {
	AddEmptyField(name string, promptDescription string, fieldType FieldType)
	AddPrefilledField(name string, value interface{})
	ProcessNextField(reqenv *base.RequestEnv)
	DoAction(reqenv *base.RequestEnv)
}

var registeredWizardActions = make(map[string]FormAction)

type Form struct {
	Fields     []*Field
	Index      int
	WizardType string

	storage StateStorage
	action  FormAction
}

func (form *Form) AddEmptyField(name string, promptDescription string, fieldType FieldType) {
	field := &Field{
		Name:              name,
		Type:              fieldType,
		PromptDescription: promptDescription,
	}
	form.Fields = append(form.Fields, field)
}

func (form *Form) AddPrefilledField(name string, value interface{}) {
	field := &Field{Name: name, Data: value}
	form.Fields = append(form.Fields, field)
}

func (form *Form) ProcessNextField(reqenv *base.RequestEnv) {
start:
	if form.Index >= len(form.Fields) {
		form.DoAction(reqenv)
		return
	}

	for form.Fields[form.Index].Data != nil {
		form.Index++
	}

	currentField := form.Fields[form.Index]
	if currentField.Data == nil {
		if currentField.WasRequested {
			currentField.Data = currentField.extractor(reqenv.Message)
			form.Index++
			goto start
		} else {
			currentField.AskUser(reqenv.Bot, reqenv.Message)
			currentField.WasRequested = true
		}
	}

	err := form.storage.SaveState(reqenv.Message.From.ID, form)
	if err != nil {
		log.Error(err)
	}
}

func (form *Form) DoAction(reqenv *base.RequestEnv) {
	if form.action == nil {
		reqenv.Reply(reqenv.Lang.Tr("wizard.errors.state.missing"))
		return
	}
	form.action(reqenv, form.Fields)
}

func (form *Form) PopulateRestored(reqenv *base.RequestEnv, storage StateStorage) {
	form.storage = storage
	form.action = restoreWizardAction(form)
	form.Fields[form.Index].restoreExtractor(reqenv.Message)
}

func NewWizard(name string, fields int, storage StateStorage, action FormAction) Wizard {
	if registeredWizardActions[name] == nil {
		registeredWizardActions[name] = action
	}
	return &Form{
		storage:    storage,
		Fields:     make(Fields, 0, fields),
		WizardType: name,
	}
}

func PopulateWizardActions(actions map[string]FormAction) bool {
	if len(registeredWizardActions) == 0 {
		registeredWizardActions = actions
		return true
	} else {
		return false
	}
}

func restoreWizardAction(wizard *Form) FormAction {
	return registeredWizardActions[wizard.WizardType]
}
