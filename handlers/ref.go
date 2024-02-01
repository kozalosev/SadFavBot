package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	RefStatusTrPrefix                 = "commands.ref.status."
	RefStatusSuccess                  = RefStatusTrPrefix + StatusSuccess
	RefStatusFailure                  = RefStatusTrPrefix + StatusFailure
	RefStatusNoRows                   = RefStatusTrPrefix + StatusNoRows
	RefFieldAliasesOrPackagesPromptTr = "commands.ref.fields.object"
)

type RefHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	aliasService *repo.AliasService
}

func NewRefHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *RefHandler {
	h := &RefHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		aliasService: repo.NewAliasService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *RefHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *RefHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.refAction)
	desc.AddField(FieldObject, RefFieldAliasesOrPackagesPromptTr)
	return desc
}

func (*RefHandler) GetCommands() []string {
	return refCommands
}

func (*RefHandler) GetScopes() []base.CommandScope {
	return commandScopePrivateAndGroupChats
}

func (handler *RefHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 1)
	w.AddEmptyField(FieldObject, wizard.Auto)
	if msg.ReplyToMessage != nil {
		if f, ok := w.(*wizard.Form); ok {
			replyMessage := msg.ReplyToMessage
			replyMessage.From = msg.From

			f.PopulateRestored(replyMessage, handler.GetWizardEnv())
			f.Fields.FindField(FieldObject).WasRequested = true
			w.ProcessNextField(reqenv, replyMessage)
			return
		}
	}
	if msg.Chat.IsPrivate() {
		w.ProcessNextField(reqenv, msg)
	}
}

func (handler *RefHandler) refAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	var (
		aliases []string
		err     error
	)
	objField := fields.FindField(FieldObject)
	switch objField.Type {
	case wizard.Text:
		txt := objField.Data.(wizard.Txt)
		aliases, err = handler.aliasService.ListByText(msg.From.ID, &txt)
	case wizard.Location:
		loc := objField.Data.(wizard.LocData)
		aliases, err = handler.aliasService.ListByLocation(msg.From.ID, &loc)
	default:
		file := objField.Data.(wizard.File)
		aliases, err = handler.aliasService.ListByFile(msg.From.ID, &file)
	}

	replyWith := possiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
	if err != nil {
		log.WithField(logconst.FieldHandler, "RefHandler").
			WithField(logconst.FieldMethod, "refAction").
			WithField(logconst.FieldCalledObject, "AliasService").
			Error(err)
		replyWith(RefStatusFailure)
		return
	} else if len(aliases) == 0 {
		replyWith(RefStatusNoRows)
	} else {
		title := reqenv.Lang.Tr(RefStatusSuccess)
		text := title + "\n\n" + LinePrefix + strings.Join(aliases, "\n"+LinePrefix)
		replyPossiblySelfDestroying(handler.appenv, msg, text, []tgbotapi.InlineKeyboardButton{})
	}
}
