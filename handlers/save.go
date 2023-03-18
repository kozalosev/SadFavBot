package handlers

import (
	"context"
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
	"strings"
)

const (
	SaveFieldsTrPrefix  = "commands.save.fields."
	SaveStatusTrPrefix  = "commands.save.status."
	SaveStatusSuccess   = SaveStatusTrPrefix + StatusSuccess
	SaveStatusFailure   = SaveStatusTrPrefix + StatusFailure
	SaveStatusDuplicate = SaveStatusTrPrefix + "duplicate"

	SaveStatusErrorForbiddenSymbolsInAlias = SaveFieldsTrPrefix + FieldAlias + FieldValidationErrorTrInfix + "forbidden.symbols"

	MaxAliasLen = 128
	MaxTextLen  = 4096
	ReservedSymbols = reservedSymbolsForMessage + "\n"
	reservedSymbolsForMessage = "•@|{}[]"
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
			template := lc.Tr(SaveFieldsTrPrefix + FieldAlias + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return verifyNoReservedSymbols(msg.Text, lc, SaveStatusErrorForbiddenSymbolsInAlias)
	}

	objDesc := desc.AddField(FieldObject, SaveFieldsTrPrefix+FieldObject)
	objDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxTextLen {
			template := lc.Tr(SaveFieldsTrPrefix + FieldObject + FieldMaxLengthErrorTrSuffix)
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
		if err := verifyNoReservedSymbols(title, reqenv.Lang, SaveStatusErrorForbiddenSymbolsInAlias); err != nil {
			reqenv.Bot.ReplyWithMarkdown(msg, err.Error())
			wizardForm.AddEmptyField(FieldAlias, wizard.Text)
		} else {
			wizardForm.AddPrefilledField(FieldAlias, title)
		}
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
		res, err = saveText(reqenv.Ctx, reqenv.Database, uid, itemValues.Alias, itemValues.Text)
	} else {
		res, err = saveFile(reqenv.Ctx, reqenv.Database, uid, itemValues.Alias, itemValues.Type, *itemValues.File)
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

func saveText(ctx context.Context, db *sql.DB, uid int64, alias, text string) (sql.Result, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	aliasID, err := saveAliasToSeparateTable(tx, alias)
	textID, err := saveTextToSeparateTable(tx, text)
	if err != nil {
		return nil, err
	}
	res, err := tx.Exec("INSERT INTO items (uid, type, alias, text) VALUES ($1, $2, " +
		"CASE WHEN ($3 > 0) THEN $3 ELSE (SELECT id FROM aliases WHERE name = $4) END, " +
		"CASE WHEN ($5 > 0) THEN $5 ELSE (SELECT id FROM texts WHERE text = $6) END)",
		uid, wizard.Text, aliasID, alias, textID, text)
	if err != nil {
		return nil, err
	}
	return res, tx.Commit()
}

func saveFile(ctx context.Context, db *sql.DB, uid int64, alias string, fileType wizard.FieldType, file wizard.File) (sql.Result, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	id, err := saveAliasToSeparateTable(tx, alias)
	if err != nil {
		return nil, err
	}
	res, err := tx.Exec("INSERT INTO items (uid, type, alias, file_id, file_unique_id) VALUES ($1, $2, CASE WHEN ($3 > 0) THEN $3 ELSE (SELECT id FROM aliases WHERE name = $4) END, $5, $6)",
		uid, fileType, id, alias, file.ID, file.UniqueID)
	if err != nil {
		return nil, err
	}
	return res, tx.Commit()
}

func saveAliasToSeparateTable(tx *sql.Tx, alias string) (int64, error) {
	var id int64
	if err := tx.QueryRow("INSERT INTO aliases(name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id", alias).Scan(&id); err == nil {
		return id, nil
	} else if err == sql.ErrNoRows {
		return 0, nil
	} else {
		return 0, err
	}
}

func saveTextToSeparateTable(tx *sql.Tx, text string) (int64, error) {
	var id int64
	if err := tx.QueryRow("INSERT INTO texts(text) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id", text).Scan(&id); err == nil {
		return id, nil
	} else if err == sql.ErrNoRows {
		return 0, nil
	} else {
		return 0, err
	}
}

func verifyNoReservedSymbols(text string, lc *loc.Context, errTemplateName string) error {
	if strings.ContainsAny(text, ReservedSymbols) {
		template := lc.Tr(errTemplateName)
		return errors.New(fmt.Sprintf(template, reservedSymbolsForMessage))
	} else {
		return nil
	}
}
