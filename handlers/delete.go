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
	DeleteFieldsTrPrefix = "commands.delete.fields."
	DeleteStatusTrPrefix = "commands.delete.status."
	DeleteStatusSuccess  = DeleteStatusTrPrefix + StatusSuccess
	DeleteStatusFailure  = DeleteStatusTrPrefix + StatusFailure
	DeleteStatusNoRows   = DeleteStatusTrPrefix + StatusNoRows
	Yes                  = "👍"
	No                   = "👎"
	SelectObjectBtnTr    = "commands.delete.button.select.object"
)

type DeleteHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	favService   *repo.FavService
	aliasService *repo.AliasService
}

func NewDeleteHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *DeleteHandler {
	h := &DeleteHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		favService:   repo.NewFavsService(appenv),
		aliasService: repo.NewAliasService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *DeleteHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *DeleteHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.deleteFormAction)

	aliasDesc := desc.AddField(FieldAlias, DeleteFieldsTrPrefix+FieldAlias)
	aliasDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len([]rune(msg.Text)) > MaxAliasLen {
			template := lc.Tr(DeleteFieldsTrPrefix + FieldAlias + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, MaxAliasLen))
		}
		return nil
	}
	aliasDesc.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		aliases, err := handler.aliasService.List(msg.From.ID)
		if err != nil {
			log.WithField(logconst.FieldHandler, "DeleteHandler").
				WithField(logconst.FieldFunc, "ReplyKeyboardBuilder").
				WithField(logconst.FieldCalledObject, "AliasService").
				WithField(logconst.FieldCalledMethod, "List").
				Error(err)
		}
		return aliases
	}
	aliasDesc.DisableKeyboardValidation = true

	delAllDesc := desc.AddField(FieldDeleteAll, DeleteFieldsTrPrefix+FieldDeleteAll)
	delAllDesc.InlineKeyboardAnswers = []string{Yes, No}

	objDesc := desc.AddField(FieldObject, DeleteFieldsTrPrefix+FieldObject)
	objDesc.SkipIf = &wizard.SkipOnFieldValue{
		Name:  FieldDeleteAll,
		Value: Yes,
	}
	objDesc.InlineKeyboardAnswers = []string{SelectObjectBtnTr}
	objDesc.DisableKeyboardValidation = true
	objDesc.InlineButtonCustomizer(SelectObjectBtnTr, func(btn *tgbotapi.InlineKeyboardButton, f *wizard.Field) {
		query := f.Form.Fields.FindField(FieldAlias).Data.(wizard.Txt).Value
		btn.SwitchInlineQueryCurrentChat = &query
	})

	return desc
}

func (*DeleteHandler) GetCommands() []string {
	return deleteCommands
}

func (*DeleteHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateAndGroupChats
}

func (handler *DeleteHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 3)
	arg := msg.CommandArguments()

	if len(arg) > 0 {
		w.AddPrefilledField(FieldAlias, arg)

		if msg.ReplyToMessage != nil {
			if f, ok := w.(*wizard.Form); ok {
				w.AddPrefilledField(FieldDeleteAll, No)
				w.AddEmptyField(FieldObject, wizard.Auto)

				replyMessage := msg.ReplyToMessage
				replyMessage.From = msg.From

				f.Index = 2
				f.PopulateRestored(replyMessage, handler.GetWizardEnv())
				f.Fields.FindField(FieldObject).WasRequested = true
				w.ProcessNextField(reqenv, replyMessage)
				return
			}
		}
	} else {
		w.AddEmptyField(FieldAlias, wizard.Text)
	}

	// only short-handed forms of commands, running in one command without the use of wizards, are supported in group chats
	if common.IsGroup(msg.Chat) {
		return
	}

	w.AddEmptyField(FieldDeleteAll, wizard.Text)
	w.AddEmptyField(FieldObject, wizard.Auto)

	w.ProcessNextField(reqenv, msg)
}

func (handler *DeleteHandler) deleteFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)

	var (
		deleteAll      bool
		deleteAllField = fields.FindField(FieldDeleteAll).Data
	)
	if data, ok := deleteAllField.(wizard.Txt); ok {
		deleteAll = data.Value == Yes
	} else {
		log.WithField(logconst.FieldHandler, "DeleteHandler").
			WithField(logconst.FieldMethod, "deleteFormAction").
			Errorf("Invalid type of Data: %T %+v", deleteAllField, deleteAllField)
		replyWith(DeleteStatusFailure)
		return
	}

	alias, fav := common.ExtractFavInfo(fields)
	if len(alias) == 0 {
		replyWith(DeleteStatusFailure)
		return
	}

	var (
		res repo.RowsAffectedAware
		err error
	)
	if deleteAll {
		res, err = handler.favService.DeleteByAlias(uid, alias)
	} else {
		res, err = handler.favService.DeleteFav(uid, alias, fav)
	}

	if err != nil {
		log.WithField(logconst.FieldHandler, "DeleteHandler").
			WithField(logconst.FieldMethod, "deleteFormAction").
			WithField(logconst.FieldCalledObject, "FavService").
			Error(err)
		replyWith(DeleteStatusFailure)
	} else {
		if res.RowsAffected() > 0 {
			replyWith(DeleteStatusSuccess)
		} else {
			replyWith(DeleteStatusNoRows)
		}
	}
}
