package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
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
	common.GroupCommandTrait

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
	linkNameDesc := desc.AddField(FieldName, RmLinkFieldTrPrefix+FieldName)
	linkNameDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len([]rune(msg.Text)) > MaxAliasLen {
			template := lc.Tr(DeleteFieldsTrPrefix + FieldName + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, MaxAliasLen))
		}
		return nil
	}
	linkNameDesc.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		aliases, err := handler.linkService.List(msg.From.ID)
		if err != nil {
			log.WithField(logconst.FieldHandler, "RemoveLinkHandler").
				WithField(logconst.FieldFunc, "ReplyKeyboardBuilder").
				WithField(logconst.FieldCalledObject, "LinkService").
				WithField(logconst.FieldCalledMethod, "List").
				Error(err)
		}
		return aliases
	}
	return desc
}

func (*RemoveLinkHandler) GetCommands() []string {
	return rmLinkCommands
}

func (*RemoveLinkHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateAndGroupChats
}

func (handler *RemoveLinkHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 1)
	if name := msg.CommandArguments(); len(name) > 0 {
		w.AddPrefilledField(FieldName, name)
	} else if msg.ReplyToMessage != nil && len(msg.ReplyToMessage.Text) > 0 {
		w.AddPrefilledField(FieldName, msg.ReplyToMessage.Text)
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
	}

	// only short-handed forms of commands, running in one command without the use of wizards, are supported in group chats
	if common.IsGroup(&msg.Chat) && !w.AllRequiredFieldsFilled() {
		return
	}

	w.ProcessNextField(reqenv, msg)
}

func (handler *RemoveLinkHandler) rmLinkAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	name := fields.FindField(FieldName).Data.(wizard.Txt).Value

	err := handler.linkService.Delete(uid, name)

	reply := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
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
