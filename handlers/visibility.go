package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	AliasVisibilityFieldTrPrefix = "commands.visibility.fields."
	FieldChange                  = "change"

	AliasVisibilityStatusTrPrefix = "commands.visibility.status."
	AliasVisibilityStatusFailure  = AliasVisibilityStatusTrPrefix + StatusFailure
	AliasVisibilityStatusNoRows   = AliasVisibilityStatusTrPrefix + "no.rows"

	ExcludeAction = "exclude"
	RevealAction  = "reveal"
)

type AliasVisibilityHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	aliasService *repo.AliasService
}

func NewAliasVisibilityHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *AliasVisibilityHandler {
	h := &AliasVisibilityHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		aliasService: repo.NewAliasService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *AliasVisibilityHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *AliasVisibilityHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.action)

	aliasField := desc.AddField(FieldAlias, AliasVisibilityFieldTrPrefix+FieldAlias)
	aliasField.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		aliases, err := handler.aliasService.ListForFavsOnly(msg.From.ID)
		if err != nil {
			log.WithField(logconst.FieldHandler, "AliasVisibilityHandler").
				WithField(logconst.FieldFunc, "ReplyKeyboardBuilder").
				WithField(logconst.FieldCalledObject, "AliasService").
				WithField(logconst.FieldCalledMethod, "ListForFavsOnly").
				Error(err)
		}
		return aliases
	}
	aliasField.DisableKeyboardValidation = true // since hidden aliases is not present in the list

	changeField := desc.AddField(FieldChange, AliasVisibilityFieldTrPrefix+FieldChange)
	changeField.InlineKeyboardAnswers = []string{ExcludeAction, RevealAction}

	return desc
}

func (*AliasVisibilityHandler) GetCommands() []string {
	return visibilityCommands
}

func (handler *AliasVisibilityHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 2)
	if alias := base.GetCommandArgument(msg); len(alias) > 0 {
		w.AddPrefilledField(FieldAlias, alias)
	} else {
		w.AddEmptyField(FieldAlias, wizard.Text)
	}
	w.AddEmptyField(FieldChange, wizard.Text)
	w.ProcessNextField(reqenv, msg)
}

func (handler *AliasVisibilityHandler) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	alias := fields.FindField(FieldAlias).Data.(string)
	visibility := fields.FindField(FieldChange).Data

	var err error
	switch visibility {
	case ExcludeAction:
		err = handler.aliasService.Hide(uid, alias)
	case RevealAction:
		err = handler.aliasService.Reveal(uid, alias)
	default:
		err = fmt.Errorf("unknown visibility: %s", visibility)
	}

	reply := base.NewReplier(handler.appenv, reqenv, msg)
	if errors.Is(err, repo.NoRowsWereAffected) {
		reply(AliasVisibilityStatusNoRows)
	} else if err != nil {
		log.WithField(logconst.FieldHandler, "AliasVisibilityHandler").
			WithField(logconst.FieldMethod, "action").
			Error(err)
		reply(AliasVisibilityStatusFailure)
	} else {
		reply(StatusSuccess)
	}
}
