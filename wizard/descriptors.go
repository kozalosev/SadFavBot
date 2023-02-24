package wizard

import (
	"fmt"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/thoas/go-funk"
)

type FormDescriptor struct {
	action FormAction
	fields map[string]*FieldDescriptor
}

type FieldDescriptor struct {
	Validator            FieldValidator
	SkipIf               SkipCondition
	ReplyKeyboardAnswers []string

	promptDescription string
	formDescriptor    *FormDescriptor
}

var registeredWizardDescriptors = make(map[string]*FormDescriptor)

func NewWizardDescriptor(action FormAction) *FormDescriptor {
	return &FormDescriptor{action: action, fields: make(map[string]*FieldDescriptor)}
}

func (descriptor *FormDescriptor) AddField(name, promptDescriptionOrTrKey string) *FieldDescriptor {
	fieldDescriptor := &FieldDescriptor{
		promptDescription: promptDescriptionOrTrKey,
		formDescriptor:    descriptor,
	}
	descriptor.fields[name] = fieldDescriptor
	return fieldDescriptor
}

func PopulateWizardDescriptors(handlers []base.MessageHandler) bool {
	if len(registeredWizardDescriptors) > 0 {
		return false
	}

	filteredHandlers := funk.Filter(handlers, func(h base.MessageHandler) bool {
		_, ok := h.(WizardMessageHandler)
		return ok
	}).([]base.MessageHandler)
	wizardHandlers := funk.Map(filteredHandlers, func(wh base.MessageHandler) WizardMessageHandler { return wh.(WizardMessageHandler) })

	descriptorsMap := funk.Map(wizardHandlers, func(wh WizardMessageHandler) (string, *FormDescriptor) {
		return wh.GetWizardName(), wh.GetWizardDescriptor()
	}).(map[string]*FormDescriptor)

	registeredWizardDescriptors = descriptorsMap
	return true
}

func (descriptor *FormDescriptor) findFieldDescriptor(name string) *FieldDescriptor {
	fieldDesc, ok := descriptor.fields[name]
	if ok {
		return fieldDesc
	} else {
		panic(fmt.Sprintf("No descriptor was found for field '%s'", name))
	}
}

func findFormDescriptor(name string) *FormDescriptor {
	desc, ok := registeredWizardDescriptors[name]
	if ok {
		return desc
	} else {
		return nil
	}
}
