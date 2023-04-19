package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/wizard"
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
	StateStorage wizard.StateStorage
}

func (handler DeleteHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler DeleteHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(deleteFormAction)

	aliasDesc := desc.AddField(FieldAlias, DeleteFieldsTrPrefix+FieldAlias)
	aliasDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxAliasLen {
			template := lc.Tr(DeleteFieldsTrPrefix + FieldAlias + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return nil
	}
	aliasDesc.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		aliasService := repo.NewAliasService(reqenv)
		aliases, err := aliasService.List(msg.From.ID)
		if err != nil {
			log.Error(err)
		}
		return aliases
	}

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
		query := f.Form.Fields.FindField(FieldAlias).Data.(string)
		btn.SwitchInlineQueryCurrentChat = &query
	})

	return desc
}

func (DeleteHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "delete" || msg.Command() == "del"
}

func (handler DeleteHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 3)
	arg := base.GetCommandArgument(msg)

	if len(arg) > 0 {
		w.AddPrefilledField(FieldAlias, arg)
	} else {
		w.AddEmptyField(FieldAlias, wizard.Text)
	}
	w.AddEmptyField(FieldDeleteAll, wizard.Text)
	w.AddEmptyField(FieldObject, wizard.Auto)

	w.ProcessNextField(reqenv, msg)
}

func deleteFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	deleteAll := fields.FindField(FieldDeleteAll).Data == Yes
	alias, fav := extractFavInfo(fields)

	replyWith := replierFactory(reqenv, msg)
	if len(alias) == 0 {
		replyWith(DeleteStatusFailure)
		return
	}

	var (
		res        repo.RowsAffectedAware
		err        error
		favService = repo.NewFavsService(reqenv)
	)
	if deleteAll {
		res, err = favService.DeleteByAlias(uid, alias)
	} else {
		res, err = favService.DeleteFav(uid, alias, fav)
	}

	if err != nil {
		log.Errorln(err.Error())
		replyWith(DeleteStatusFailure)
	} else {
		if res.RowsAffected() > 0 {
			replyWith(DeleteStatusSuccess)
		} else {
			replyWith(DeleteStatusNoRows)
		}
	}
}
