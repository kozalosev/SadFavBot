package wizard

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jinzhu/copier"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestCallbackQueryHandler(t *testing.T) {
	msg := &tgbotapi.Message{
		Text:      "not" + TestValue,
		Chat:      &tgbotapi.Chat{ID: TestID},
		MessageID: TestID,
		From:      &tgbotapi.User{ID: TestID},
	}
	msg.ReplyToMessage = msg
	query := &tgbotapi.CallbackQuery{
		ID:      strconv.Itoa(TestID),
		From:    msg.From,
		Message: msg,
	}
	reqenv := &base.RequestEnv{
		Bot:  &base.BotAPI{DummyMode: true},
		Lang: loc.NewPool("en").GetContext("en"),
	}

	actionFlagCont := &flagContainer{}
	storage := inMemoryStorage{storage: make(map[int64]Wizard, 1)}
	handler := testHandlerWithAction{stateStorage: storage, actionWasRunFlag: actionFlagCont}
	clearRegisteredDescriptors()
	PopulateWizardDescriptors([]base.MessageHandler{handler})

	wizard := NewWizard(handler, 2)
	wizard.AddEmptyField(TestName, Text)
	wizard.AddEmptyField(TestName2, Text)
	form := wizard.(*Form)

	_ = storage.SaveState(TestID, form)

	query.Data = fmt.Sprintf("%s%s:%s", callbackDataFieldPrefix, TestName, TestValue)
	CallbackQueryHandler(reqenv, query, storage)
	_ = storage.GetCurrentState(TestID, form)

	assert.Equal(t, 1, form.Index)
	assert.False(t, form.Fields[0].WasRequested)
	assert.True(t, form.Fields[1].WasRequested)
	assert.Equal(t, TestValue, form.Fields[0].Data)
	assert.Nil(t, form.Fields[1].Data)
	assert.False(t, actionFlagCont.flag)

	query.Data = fmt.Sprintf("%s%s:%s", callbackDataFieldPrefix, TestName2, TestValue)
	CallbackQueryHandler(reqenv, query, storage)
	_ = storage.GetCurrentState(TestID, form)

	assert.Equal(t, 2, form.Index)
	assert.Equal(t, TestValue, form.Fields[1].Data)
	assert.True(t, actionFlagCont.flag)
}

type inMemoryStorage struct {
	storage map[int64]Wizard
}

func (i inMemoryStorage) GetCurrentState(uid int64, dest Wizard) error {
	return copier.Copy(dest, i.storage[uid])
}

func (i inMemoryStorage) SaveState(uid int64, wizard Wizard) error {
	i.storage[uid] = wizard
	return nil
}

func (i inMemoryStorage) DeleteState(uid int64) error {
	delete(i.storage, uid)
	return nil
}

func (i inMemoryStorage) Close() error {
	return nil
}
