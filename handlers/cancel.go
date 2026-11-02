package handlers

import (
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const SuccessTr = "success"

type CancelHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage
}

func NewCancelHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *CancelHandler {
	h := &CancelHandler{
		appEnv:       appenv,
		stateStorage: stateStorage,
	}
	h.HandlerRefForTrait = h
	return h
}

func (*CancelHandler) GetCommands() []string {
	return cancelCommands
}

func (c *CancelHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	err := c.stateStorage.DeleteState(msg.From.ID)
	var answer string
	if err != nil {
		answer = reqenv.Lang.Tr(err.Error())
	} else {
		answer = reqenv.Lang.Tr(SuccessTr)
	}
	// reply keyboards are not shown in groups, so there is nothing to remove there
	if base.IsGroupChat(&msg.Chat) {
		common.ReplyPossiblySelfDestroying(c.appEnv, reqenv.Lang, msg, answer, base.NoOpCustomizer)
		return
	}
	c.appEnv.Bot.ReplyWithMessageCustomizer(msg, answer, func(m *tgbotapi.MessageConfig) {
		m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	})
}
