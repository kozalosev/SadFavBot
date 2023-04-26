package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const SuccessTr = "success"

type CancelHandler struct {
	base.CommandHandlerTrait

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
	c.appEnv.Bot.Reply(msg, answer)
}
