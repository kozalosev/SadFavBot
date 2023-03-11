package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
)

const SuccessTr = "success"

type CancelHandler struct {
	StateStorage wizard.StateStorage
}

func (CancelHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "cancel"
}

func (c CancelHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	err := c.StateStorage.DeleteState(msg.From.ID)
	var answer string
	if err != nil {
		answer = reqenv.Lang.Tr(err.Error())
	} else {
		answer = reqenv.Lang.Tr(SuccessTr)
	}
	reqenv.Bot.Reply(msg, answer)
}
