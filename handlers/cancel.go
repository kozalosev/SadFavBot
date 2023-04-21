package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
)

const SuccessTr = "success"

type CancelHandler struct {
	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage
}

func NewCancelHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *CancelHandler {
	return &CancelHandler{
		appEnv:       appenv,
		stateStorage: stateStorage,
	}
}

func (*CancelHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "cancel"
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
