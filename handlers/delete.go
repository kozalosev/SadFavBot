package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
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
)

type DeleteHandler struct {
	StateStorage wizard.StateStorage
}

func (DeleteHandler) GetWizardName() string                              { return "DeleteWizard" }
func (handler DeleteHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler DeleteHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(deleteFormAction)

	aliasDesc := desc.AddField(FieldAlias, DeleteFieldsTrPrefix+FieldAlias)
	aliasDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxAliasLen {
			template := lc.Tr(DeleteFieldsTrPrefix + FieldAlias + FieldValidationErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return nil
	}

	delAllDesc := desc.AddField(FieldDeleteAll, DeleteFieldsTrPrefix+FieldDeleteAll)
	delAllDesc.InlineKeyboardAnswers = []string{Yes, No}

	objDesc := desc.AddField(FieldObject, DeleteFieldsTrPrefix+FieldObject)
	objDesc.SkipIf = &wizard.SkipOnFieldValue{
		Name:  FieldDeleteAll,
		Value: Yes,
	}

	return desc
}

func (DeleteHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "delete" || msg.Command() == "del"
}

func (handler DeleteHandler) Handle(reqenv *base.RequestEnv) {
	w := wizard.NewWizard(handler, 3)
	arg := base.GetCommandArgument(reqenv.Message)

	if len(arg) > 0 {
		w.AddPrefilledField(FieldAlias, arg)
	} else {
		w.AddEmptyField(FieldAlias, wizard.Text)
	}
	w.AddEmptyField(FieldDeleteAll, wizard.Text)
	w.AddEmptyField(FieldObject, wizard.Auto)

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
		res, err = deleteByFileID(reqenv.Database, uid, itemValues.Alias, *itemValues.File)
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
