package handlers

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	RmLinkFieldTrPrefix  = "commands.rmlink.fields."
	RmLinkStatusTrPrefix = "commands.rmlink.status."
	RmLinkStatusSuccess  = RmLinkStatusTrPrefix + StatusSuccess
	RmLinkStatusFailure  = RmLinkStatusTrPrefix + StatusFailure
	RmLinkStatusNoRows   = RmLinkStatusTrPrefix + StatusNoRows
)

type RemoveLinkHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	linkService *repo.LinkService
}

func NewRemoveLinkHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *RemoveLinkHandler {
	h := &RemoveLinkHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		linkService:  repo.NewLinkService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *RemoveLinkHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *RemoveLinkHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.rmLinkAction)
	desc.AddField(FieldName, RmLinkFieldTrPrefix+FieldName)
	return desc
}

func (*RemoveLinkHandler) GetCommands() []string {
	return rmLinkCommands
}

func (handler *RemoveLinkHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 1)
	if name := base.GetCommandArgument(msg); len(name) > 0 {
		w.AddPrefilledField(FieldName, name)
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func (handler *RemoveLinkHandler) rmLinkAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	name := fields.FindField(FieldName).Data.(wizard.Txt).Value

	err := handler.linkService.Delete(uid, name)

	reply := base.NewReplier(handler.appenv, reqenv, msg)
	if errors.Is(err, repo.NoRowsWereAffected) {
		reply(RmLinkStatusNoRows)
	} else if err != nil {
		log.WithField(logconst.FieldHandler, "RemoveLinkHandler").
			WithField(logconst.FieldMethod, "rmLinkAction").
			WithField(logconst.FieldCalledObject, "LinkService").
			WithField(logconst.FieldCalledMethod, "Delete").
			Error(err)
		reply(RmLinkStatusFailure)
	} else {
		reply(RmLinkStatusSuccess)
	}
}
