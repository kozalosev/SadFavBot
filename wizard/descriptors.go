package wizard

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/thoas/go-funk"
)

// InlineButtonCustomizer is a function that allows you to customize the inline button generated for a field prompt.
// https://core.telegram.org/bots/api#inlinekeyboardbutton
type InlineButtonCustomizer func(btn *tgbotapi.InlineKeyboardButton, f *Field)

// ReplyKeyboardBuilder is used to generate variants for the reply keyboard, since I use it for fields when the user
// must choose the option from a result set fetched from the database.
type ReplyKeyboardBuilder func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string

// FormDescriptor is the description of a wizard, describing all non-storable parameters.
// Use [NewWizardDescriptor] to create one.
type FormDescriptor struct {
	action FormAction
	fields map[string]*FieldDescriptor
}

// FieldDescriptor is the description of a concrete field of the form, describing all non-storable parameters.
// Use [FormDescriptor.AddField] to create one and attach to a [FormDescriptor] instance.
type FieldDescriptor struct {
	Validator FieldValidator

	// if this condition is true, the field will be skipped
	SkipIf SkipCondition

	// keyboard options; you can attach either a reply keyboard or inline one, but not both
	ReplyKeyboardBuilder      ReplyKeyboardBuilder
	InlineKeyboardAnswers     []string
	DisableKeyboardValidation bool

	// this text will be used to ask the user for the field value
	promptDescription string

	formDescriptor          *FormDescriptor
	inlineButtonCustomizers map[string]InlineButtonCustomizer
}

// in-memory storage of all descriptors; use [PopulateWizardDescriptors] to register them at startup
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

// InlineButtonCustomizer is a method for modifying the inline button generated for a specific option.
// By default, an inline button with callback_data is created. By using this type of customizer, you're able to change this behavior.
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

// PopulateWizardDescriptors fills in the map that should be initialized at startup time to prevent the user from
// receiving the "wizard.errors.state.missing" message.
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
