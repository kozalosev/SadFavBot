package handlers

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	DeleteFieldsTrPrefix = "commands.delete.fields."
	DeleteStatusTrPrefix = "commands.delete.status."
	Yes                  = "👍"
	No                   = "👎"
)

type DeleteHandler struct {
	StateStorage wizard.StateStorage
}

func (DeleteHandler) GetWizardName() string              { return "DeleteWizard" }
func (DeleteHandler) GetWizardAction() wizard.FormAction { return deleteFormAction }

func (handler DeleteHandler) GetWizardStateStorage() wizard.StateStorage {
	return handler.StateStorage
}

func (DeleteHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "delete" || msg.Command() == "del"
}

func (handler DeleteHandler) Handle(reqenv *base.RequestEnv) {
	w := wizard.NewWizard(handler, 3)
	arg := base.GetCommandArgument(reqenv.Message)
	skipCondition, err := wizard.WrapCondition(&wizard.SkipOnFieldValue{
		Name:  "deleteAll",
		Value: Yes,
	})
	if err != nil {
		log.Errorln(err)
		reqenv.Reply(reqenv.Lang.Tr("errors.unknown"))
	}

	if len(arg) > 0 {
		w.AddPrefilledField(FieldAlias, arg)
	} else {
		w.AddEmptyField(FieldAlias, reqenv.Lang.Tr(DeleteFieldsTrPrefix+FieldAlias), wizard.Text)
	}
	deleteAllField := w.AddEmptyField(FieldDeleteAll, reqenv.Lang.Tr(DeleteFieldsTrPrefix+FieldDeleteAll), wizard.Text)
	deleteAllField.InlineKeyboardAnswers = []string{Yes, No}
	objectField := w.AddEmptyField(FieldObject, reqenv.Lang.Tr(DeleteFieldsTrPrefix+FieldObject), wizard.Auto)
	objectField.SkipIf = skipCondition
	w.ProcessNextField(reqenv)
}

func deleteFormAction(reqenv *base.RequestEnv, fields wizard.Fields) {
	uid := reqenv.Message.From.ID
	alias := fields.FindField(FieldAlias).Data
	deleteAll := fields.FindField(FieldDeleteAll).Data == Yes
	object := fields.FindField(FieldObject).Data
	file, ok := object.(wizard.File)
	if !ok {
		log.Errorf("Invalid type: File was expected but '%T %+v' is got", object, object)
		reqenv.Reply(reqenv.Lang.Tr(DeleteStatusTrPrefix + StatusFailure))
	}

	var (
		res sql.Result
		err error
	)
	if deleteAll {
		log.Infof("Deletion of items with uid '%d' and alias '%s'", uid, alias)
		res, err = reqenv.Database.Exec("DELETE FROM item WHERE uid = $1 AND alias = $2",
			uid, alias)
	} else {
		log.Infof("Deletion of items with uid '%d', alias '%s' and file_id '%s'", uid, alias, file.FileUniqueID)
		res, err = reqenv.Database.Exec("DELETE FROM item WHERE uid = $1 AND alias = $2 AND file_unique_id = $3",
			uid, alias, file.FileUniqueID)
	}
	if err != nil {
		log.Errorln(err.Error())
		reqenv.Reply(reqenv.Lang.Tr(DeleteStatusTrPrefix + StatusFailure))
	} else {
		var rowsAffected int64
		if rowsAffected, err = res.RowsAffected(); err != nil {
			log.Errorln(err)
			rowsAffected = -1 // logs but ignores
		}
		if rowsAffected == 0 {
			reqenv.Reply(reqenv.Lang.Tr(DeleteStatusTrPrefix + "no.rows"))
		} else {
			reqenv.Reply(reqenv.Lang.Tr(DeleteStatusTrPrefix + StatusSuccess))
		}
	}
}
