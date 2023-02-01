package wizard

import (
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type FormAction func(request *base.RequestEnv, fields Fields)

type Wizard interface {
	AddEmptyField(name, promptDescription string, fieldType FieldType) *Field
	AddPrefilledField(name string, value interface{})
	ProcessNextField(reqenv *base.RequestEnv)
	DoAction(reqenv *base.RequestEnv)
	PopulateRestored(reqenv *base.RequestEnv, storage StateStorage)
}

//goland:noinspection GoNameStartsWithPackageName
type WizardMessageHandler interface {
	base.MessageHandler

	GetWizardName() string
	GetWizardAction() FormAction
	GetWizardStateStorage() StateStorage
}

var registeredWizardActions = make(map[string]FormAction)

type Form struct {
	Fields     Fields
	Index      int
	WizardType string

	storage StateStorage
	action  FormAction
}

func (form *Form) AddEmptyField(name, promptDescription string, fieldType FieldType) *Field {
	field := &Field{
		Name:              name,
		Type:              fieldType,
		PromptDescription: promptDescription,
	}
	form.Fields = append(form.Fields, field)
	return field
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
		currentField.Data = currentField.extractor(reqenv.Message)
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

func NewWizard(handler WizardMessageHandler, fields int) Wizard {
	name := handler.GetWizardName()
	if registeredWizardActions[name] == nil {
		registeredWizardActions[name] = handler.GetWizardAction()
	}
	return &Form{
		storage:    handler.GetWizardStateStorage(),
		Fields:     make(Fields, 0, fields),
		WizardType: name,
		action:     handler.GetWizardAction(),
	}
}

func PopulateWizardActions(handlers []base.MessageHandler) bool {
	if len(registeredWizardActions) > 0 {
		return false
	}

	filteredHandlers := funk.Filter(handlers, func(h base.MessageHandler) bool {
		_, ok := h.(WizardMessageHandler)
		return ok
	}).([]base.MessageHandler)
	wizardHandlers := funk.Map(filteredHandlers, func(wh base.MessageHandler) WizardMessageHandler { return wh.(WizardMessageHandler) })

	actionsMap := funk.Map(wizardHandlers, func(wh WizardMessageHandler) (string, FormAction) {
		return wh.GetWizardName(), wh.GetWizardAction()
	}).(map[string]FormAction)

	registeredWizardActions = actionsMap
	return true
}

func restoreWizardAction(wizard *Form) FormAction {
	return registeredWizardActions[wizard.WizardType]
}

func shouldBeSkipped(field *Field, form *Form) bool {
	c, err := UnwrapCondition(field.SkipIf)
	if err != nil {
		log.Error(err)
		return false
	}
	return c.ShouldBeSkipped(form)
}
