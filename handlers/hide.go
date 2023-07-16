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
	HideStatusTrPrefix = "commands.hide.status."
	HideStatusFailure  = HideStatusTrPrefix + StatusFailure
	HideStatusNoRows   = HideStatusTrPrefix + "no.rows"
)

type HideHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	favsService  *repo.FavService
	aliasService *repo.AliasService
}

func NewHideHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *HideHandler {
	h := &HideHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		favsService:  repo.NewFavsService(appenv),
		aliasService: repo.NewAliasService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *HideHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *HideHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.hideAction)

	aliasField := desc.AddField(FieldAlias, LinkFieldTrPrefix+FieldAlias)
	aliasField.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		aliases, err := handler.aliasService.ListForFavsOnly(msg.From.ID)
		if err != nil {
			log.WithField(logconst.FieldHandler, "HideHandler").
				WithField(logconst.FieldFunc, "ReplyKeyboardBuilder").
				WithField(logconst.FieldCalledObject, "AliasService").
				WithField(logconst.FieldCalledMethod, "ListForFavsOnly").
				Error(err)
		}
		return aliases
	}

	return desc
}

func (*HideHandler) GetCommands() []string {
	return hideCommands
}

func (handler *HideHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 1)
	if alias := base.GetCommandArgument(msg); len(alias) > 0 {
		w.AddPrefilledField(FieldAlias, alias)
	} else {
		w.AddEmptyField(FieldAlias, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func (handler *HideHandler) hideAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	alias := fields.FindField(FieldAlias).Data.(string)

	err := handler.favsService.Hide(uid, alias)

	reply := base.NewReplier(handler.appenv, reqenv, msg)
	if errors.Is(err, repo.NoRowsWereAffected) {
		reply(HideStatusNoRows)
	} else if err != nil {
		log.WithField(logconst.FieldHandler, "HideHandler").
			WithField(logconst.FieldMethod, "hideAction").
			WithField(logconst.FieldCalledObject, "FavsService").
			WithField(logconst.FieldCalledMethod, "Hide").
			Error(err)
		reply(HideStatusFailure)
	} else {
		reply(StatusSuccess)
	}
}
