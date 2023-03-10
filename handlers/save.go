package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"strconv"
)

const (
	SaveFieldsTrPrefix  = "commands.save.fields."
	SaveStatusTrPrefix  = "commands.save.status."
	SaveStatusSuccess   = SaveStatusTrPrefix + StatusSuccess
	SaveStatusFailure   = SaveStatusTrPrefix + StatusFailure
	SaveStatusDuplicate = SaveStatusTrPrefix + "duplicate"

	MaxAliasLen = 1024
	MaxTextLen  = 10000
)

var (
	maxAliasLenStr = strconv.FormatInt(MaxAliasLen, 10)
	maxTextLenStr  = strconv.FormatInt(MaxAliasLen, 10)
)

type SaveHandler struct {
	StateStorage wizard.StateStorage
}

func (SaveHandler) GetWizardName() string                              { return "SaveWizard" }
func (handler SaveHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler SaveHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(saveFormAction)

	aliasDesc := desc.AddField(FieldAlias, SaveFieldsTrPrefix+FieldAlias)
	aliasDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxAliasLen {
			template := lc.Tr(SaveFieldsTrPrefix + FieldAlias + FieldValidationErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return nil
	}

	objDesc := desc.AddField(FieldObject, SaveFieldsTrPrefix+FieldObject)
	objDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxTextLen {
			template := lc.Tr(SaveFieldsTrPrefix + FieldObject + FieldValidationErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxTextLenStr))
		}
		return nil
	}

	return desc
}

func (SaveHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "save"
}

func (handler SaveHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	wizardForm := wizard.NewWizard(handler, 2)
	title := base.GetCommandArgument(msg)
	if len(title) > 0 {
		wizardForm.AddPrefilledField(FieldAlias, title)
	} else {
		wizardForm.AddEmptyField(FieldAlias, wizard.Text)
	}
	wizardForm.AddEmptyField(FieldObject, wizard.Auto)
	wizardForm.ProcessNextField(reqenv, msg)
}

func saveFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	itemValues, ok := extractItemValues(fields)

	replyWith := replierFactory(reqenv, msg)
	if !ok {
		replyWith(SaveStatusFailure)
		return
	}

	var (
		res sql.Result
		err error
	)
	if itemValues.Type == wizard.Text {
		res, err = saveText(reqenv.Database, uid, itemValues.Alias, itemValues.Text)
	} else {
		res, err = saveFile(reqenv.Database, uid, itemValues.Alias, itemValues.Type, *itemValues.File)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			replyWith(SaveStatusDuplicate)
		} else {
			log.Errorln(err.Error())
			replyWith(SaveStatusFailure)
		}
	} else {
		if checkRowsWereAffected(res) {
			replyWith(SaveStatusSuccess)
		} else {
			log.Warning("No rows were affected!")
			replyWith(SaveStatusFailure)
		}
	}
}

func saveText(db *sql.DB, uid int64, alias, text string) (sql.Result, error) {
	return db.Exec("INSERT INTO items (uid, type, alias, text) VALUES ($1, $2, $3, $4)",
		uid, wizard.Text, alias, text)
}

func saveFile(db *sql.DB, uid int64, alias string, fileType wizard.FieldType, file wizard.File) (sql.Result, error) {
	return db.Exec("INSERT INTO items (uid, type, alias, file_id, file_unique_id) VALUES ($1, $2, $3, $4, $5)",
		uid, fileType, alias, file.ID, file.UniqueID)
}
