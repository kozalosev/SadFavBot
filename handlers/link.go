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
	"github.com/thoas/go-funk"
	"strings"
)

const (
	LinkFieldTrPrefix                     = "commands.link.fields."
	LinkStatusTrPrefix                    = "commands.link.status."
	LinkStatusSuccess                     = LinkStatusTrPrefix + StatusSuccess
	LinkStatusFailure                     = LinkStatusTrPrefix + StatusFailure
	LinkStatusDuplicate                   = LinkStatusTrPrefix + StatusDuplicate
	LinkStatusDuplicateAlias              = LinkStatusTrPrefix + StatusDuplicate + ".alias"
	LinkStatusNoAlias                     = LinkStatusTrPrefix + "no.alias"
	LinkStatusErrorForbiddenSymbolsInName = LinkFieldTrPrefix + FieldName + FieldValidationErrorTrInfix + "forbidden.symbols"
)

type LinkHandler struct {
	StateStorage wizard.StateStorage
}

func (LinkHandler) GetWizardName() string                              { return "LinkWizard" }
func (handler LinkHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler LinkHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(linkAction)

	nameField := desc.AddField(FieldName, LinkFieldTrPrefix+FieldName)
	nameField.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxAliasLen {
			template := lc.Tr(LinkFieldTrPrefix + FieldName + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return verifyNoReservedSymbols(msg.Text, lc, LinkStatusErrorForbiddenSymbolsInName)
	}

	aliasField := desc.AddField(FieldAlias, LinkFieldTrPrefix+FieldAlias)
	aliasField.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		var (
			aliases []string
			alias   string
		)
		if res, err := reqenv.Database.QueryContext(reqenv.Ctx, "SELECT DISTINCT a.name FROM items i JOIN aliases a on a.id = i.alias WHERE i.uid = $1", msg.From.ID); err == nil {
			for res.Next() {
				if err := res.Scan(&alias); err == nil {
					aliases = append(aliases, alias)
				} else {
					log.Error(err)
				}
			}
		} else {
			log.Error(err)
		}
		return aliases
	}

	return desc
}

func (handler LinkHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "link" || msg.Command() == "ln"
}

func (handler LinkHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 2)
	if name := base.GetCommandArgument(msg); len(name) > 0 {
		argParts := funk.Map(strings.Split(name, "->"), func(s string) string {
			return strings.TrimSpace(s)
		}).([]string)
		if len(argParts) == 2 {
			if len(argParts[0]) <= MaxAliasLen && verifyNoReservedSymbols(argParts[0], reqenv.Lang, LinkStatusErrorForbiddenSymbolsInName) == nil {
				w.AddPrefilledField(FieldName, argParts[0])
			} else {
				w.AddEmptyField(FieldName, wizard.Text)
			}
			w.AddPrefilledField(FieldAlias, argParts[1])
		} else {
			if len(name) <= MaxAliasLen && verifyNoReservedSymbols(name, reqenv.Lang, LinkStatusErrorForbiddenSymbolsInName) == nil {
				w.AddPrefilledField(FieldName, name)
			} else {
				w.AddEmptyField(FieldName, wizard.Text)
			}
			w.AddEmptyField(FieldAlias, wizard.Text)
		}
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
		w.AddEmptyField(FieldAlias, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func linkAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	name := fields.FindField(FieldName).Data.(string)
	refAlias := fields.FindField(FieldAlias).Data.(string)

	var (
		tx  *sql.Tx
		err error
	)
	if tx, err = reqenv.Database.BeginTx(reqenv.Ctx, nil); err == nil {
		var aliasID int
		if aliasID, err = saveAliasToSeparateTable(reqenv.Ctx, tx, name); err == nil {
			if _, err = tx.ExecContext(reqenv.Ctx, "INSERT INTO Links(uid, alias_id, linked_alias_id) VALUES ($1, "+
				"CASE WHEN ($2 > 0) THEN $2 ELSE (SELECT id FROM aliases WHERE name = $3) END, "+
				"(SELECT id FROM aliases WHERE name = $4))",
				uid, aliasID, name, refAlias); err == nil {
				err = tx.Commit()
			}
		}
	}

	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Error(err)
		}
	}

	reply := replierFactory(reqenv, msg)
	if isAttemptToInsertLinkForExistingAlias(err) {
		reply(LinkStatusDuplicateAlias)
	} else if isDuplicateConstraintViolation(err) {
		reply(LinkStatusDuplicate)
	} else if isAttemptToLinkNonExistingAlias(err) {
		reply(LinkStatusNoAlias)
	} else if err != nil {
		log.Error(err)
		reply(LinkStatusFailure)
	} else {
		reply(LinkStatusSuccess)
	}
}

func isAttemptToInsertLinkForExistingAlias(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Message == "Insertion of the link with the same name as an already existing alias is forbidden"
}

func isAttemptToLinkNonExistingAlias(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23502" && pgErr.ColumnName == "linked_alias_id"
}
