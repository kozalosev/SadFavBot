package wizard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestForm_AddEmptyField(t *testing.T) {
	wizard := NewWizard(testHandler{}, 1)
	wizard.AddEmptyField(TestName, TestPromptDesc, Text)

	form := wizard.(*Form)
	resField := form.Fields[0]

	assert.Equal(t, TestName, resField.Name)
	assert.Equal(t, TestPromptDesc, resField.PromptDescription)
	assert.Equal(t, Text, resField.Type)
	assert.Nil(t, resField.Data)
}

func TestForm_AddPrefilledField(t *testing.T) {
	wizard := NewWizard(testHandler{}, 1)
	wizard.AddPrefilledField(TestName, TestValue)

	form := wizard.(*Form)
	resField := form.Fields[0]

	assert.Equal(t, TestName, resField.Name)
	assert.Equal(t, TestValue, resField.Data)
	assert.Equal(t, FieldType(""), resField.Type)
}

func TestPopulationOfWizardActions(t *testing.T) {
	registeredWizardActions = make(map[string]FormAction)
	assert.Len(t, registeredWizardActions, 0)

	ok := PopulateWizardActions([]base.MessageHandler{testHandler{}})
	assert.True(t, ok)

	action := registeredWizardActions[TestWizardName]
	assert.Equal(t, getFuncPtr(tAction), getFuncPtr(action))

	assert.Len(t, registeredWizardActions, 1)
	ok = PopulateWizardActions([]base.MessageHandler{testHandler2{}}) // doesn't add anything if the map is not empty
	assert.False(t, ok)
	assert.Len(t, registeredWizardActions, 1)

	NewWizard(testHandler2{}, 0) // but NewWizard still registers new actions
	assert.Len(t, registeredWizardActions, 2)
}

func TestRestorationOfFunctions(t *testing.T) {
	wizard := NewWizard(testHandler{}, 1)
	wizard.AddEmptyField(TestName, TestPromptDesc, Text)
	reqenv := base.RequestEnv{Message: &tgbotapi.Message{}}
	wizard.PopulateRestored(&reqenv, nil)

	form := wizard.(*Form)
	assert.Equal(t, getFuncPtr(tAction), getFuncPtr(form.action))
	assert.Equal(t, getFuncPtr(textExtractor), getFuncPtr(form.Fields[form.Index].extractor))
}

func TestForm_ProcessNextField(t *testing.T) {
	reqenv := &base.RequestEnv{
		Bot: &base.BotAPI{DummyMode: true},
		Message: &tgbotapi.Message{
			Text:      TestValue,
			Chat:      &tgbotapi.Chat{ID: TestID},
			MessageID: TestID,
			From:      &tgbotapi.User{ID: TestID},
		},
	}
	f3Cond, err := WrapCondition(&SkipOnFieldValue{
		Name:  TestName2,
		Value: TestValue,
	})
	assert.NoError(t, err)

	actionFlagCont := &flagContainer{}
	wizard := NewWizard(testHandlerWithAction{stateStorage: fakeStorage{}, actionWasRunFlag: actionFlagCont}, 2)
	wizard.AddPrefilledField(TestName, TestValue)
	wizard.AddEmptyField(TestName2, TestPromptDesc, Text)
	f3 := wizard.AddEmptyField(TestName3, TestPromptDesc, Text)
	f3.SkipIf = f3Cond
	form := wizard.(*Form)

	assert.Equal(t, 0, form.Index)
	assert.False(t, form.Fields[0].WasRequested)
	assert.Equal(t, TestValue, form.Fields[0].Data)
	assert.False(t, form.Fields[1].WasRequested)
	assert.False(t, actionFlagCont.flag)

	form.ProcessNextField(reqenv)

	assert.Equal(t, 1, form.Index)
	assert.False(t, form.Fields[0].WasRequested)
	assert.True(t, form.Fields[1].WasRequested)
	assert.Nil(t, form.Fields[1].Data)
	assert.False(t, actionFlagCont.flag)

	form.Fields[1].extractor = textExtractor
	form.ProcessNextField(reqenv)

	assert.Equal(t, 3, form.Index)
	assert.Equal(t, TestValue, form.Fields[1].Data)
	assert.True(t, actionFlagCont.flag)
}

func tAction(_ *base.RequestEnv, fields Fields) {
	f3 := fields.FindField(TestName3)
	if f3.Data != nil {
		panic(TestName3 + " must be skipped and equals to nil!") // assertion without access to `t *testing.T`
	}
}

type testHandler struct{}

func (testHandler) CanHandle(*tgbotapi.Message) bool    { return false }
func (testHandler) Handle(*base.RequestEnv)             {}
func (testHandler) GetWizardName() string               { return TestWizardName }
func (testHandler) GetWizardAction() FormAction         { return tAction }
func (testHandler) GetWizardStateStorage() StateStorage { return nil }

type testHandler2 struct {
	testHandler
}

func (testHandler2) GetWizardName() string { return TestWizardName + "2" }

type flagContainer struct {
	flag bool
}
type testHandlerWithAction struct {
	testHandler

	stateStorage     StateStorage
	actionWasRunFlag *flagContainer
}

func (handler testHandlerWithAction) GetWizardAction() FormAction {
	return func(request *base.RequestEnv, fields Fields) {
		handler.actionWasRunFlag.flag = true
	}
}

func (handler testHandlerWithAction) GetWizardStateStorage() StateStorage {
	return handler.stateStorage
}

type fakeStorage struct{}

func (f fakeStorage) GetCurrentState(int64, Wizard) error { return nil }
func (f fakeStorage) SaveState(int64, Wizard) error       { return nil }
func (f fakeStorage) Close() error                        { return nil }
