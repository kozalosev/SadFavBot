package wizard

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/thoas/go-funk"
)

type InlineButtonCustomizer func(btn *tgbotapi.InlineKeyboardButton, f *Field)
type ReplyKeyboardBuilder func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string

type FormDescriptor struct {
	action FormAction
	fields map[string]*FieldDescriptor
}

type FieldDescriptor struct {
	Validator             FieldValidator
	SkipIf                SkipCondition
	ReplyKeyboardBuilder  ReplyKeyboardBuilder
	InlineKeyboardAnswers []string
	DisableKeyboardValidation bool

	promptDescription 		string
	formDescriptor          *FormDescriptor
	inlineButtonCustomizers map[string]InlineButtonCustomizer
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

func (descriptor *FieldDescriptor) InlineButtonCustomizer(option string, customizer InlineButtonCustomizer) bool {
	if descriptor.inlineButtonCustomizers == nil {
		descriptor.inlineButtonCustomizers = make(map[string]InlineButtonCustomizer, len(descriptor.InlineKeyboardAnswers))
	}
	if _, ok := descriptor.inlineButtonCustomizers[option]; ok {
		return false
	}
	descriptor.inlineButtonCustomizers[option] = customizer
	return true
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
