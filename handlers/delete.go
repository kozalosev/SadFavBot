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
	DeleteStatusSuccess  = DeleteStatusTrPrefix + StatusSuccess
	DeleteStatusFailure  = DeleteStatusTrPrefix + StatusFailure
	DeleteStatusNoRows   = DeleteStatusTrPrefix + "no.rows"
	UnknownError         = "errors.unknown"
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
		reqenv.Reply(reqenv.Lang.Tr(UnknownError))
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
	deleteAll := fields.FindField(FieldDeleteAll).Data == Yes
	itemValues, ok := extractItemValues(fields)

	replyWith := replierFactory(reqenv)
	if !ok {
		replyWith(DeleteStatusFailure)
		return
	}

	var (
		res sql.Result
		err error
	)
	if deleteAll {
		res, err = deleteByAlias(reqenv.Database, uid, itemValues.Alias)
	} else if itemValues.Type == wizard.Text {
		res, err = deleteByText(reqenv.Database, uid, itemValues.Alias, itemValues.Text)
	} else {
		res, err = deleteByFileID(reqenv.Database, uid, itemValues.Alias, itemValues.File)
	}
	if err != nil {
		log.Errorln(err.Error())
		replyWith(DeleteStatusFailure)
	} else {
		if checkRowsWereAffected(res) {
			replyWith(DeleteStatusSuccess)
		} else {
			replyWith(DeleteStatusNoRows)
		}
	}
}

func deleteByAlias(db *sql.DB, uid int64, alias string) (sql.Result, error) {
	log.Infof("Deletion of items with uid '%d' and alias '%s'", uid, alias)
	return db.Exec("DELETE FROM items WHERE uid = $1 AND alias = $2", uid, alias)
}

func deleteByFileID(db *sql.DB, uid int64, alias string, file wizard.File) (sql.Result, error) {
	log.Infof("Deletion of items with uid '%d', alias '%s' and file_id '%s'", uid, alias, file.UniqueID)
	return db.Exec("DELETE FROM items WHERE uid = $1 AND alias = $2 AND file_unique_id = $3", uid, alias, file.UniqueID)
}

func deleteByText(db *sql.DB, uid int64, alias, text string) (sql.Result, error) {
	log.Infof("Deletion of items with uid '%d', alias '%s' and text '%s'", uid, alias, text)
	return db.Exec("DELETE FROM items WHERE uid = $1 AND alias = $2 AND text = $3", uid, alias, text)
}
