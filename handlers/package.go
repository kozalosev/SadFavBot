package handlers

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"net/url"
	"strings"
)

const (
	PackageFieldsTrPrefix = "commands.package.fields."
	PackageStatusTrPrefix = "commands.package.status."
	PackageStatusCreationSuccess  = PackageStatusTrPrefix + StatusSuccess + ".creation"
	PackageStatusDeletionSuccess  = PackageStatusTrPrefix + StatusSuccess + ".deletion"
	PackageStatusFailure   = PackageStatusTrPrefix + StatusFailure
	PackageStatusDuplicate = PackageStatusTrPrefix + StatusDuplicate
	PackageStatusNoRows    = PackageStatusTrPrefix + StatusNoRows

	PackageStatusErrorForbiddenSymbolsInName = PackageFieldsTrPrefix + FieldName + FieldValidationErrorTrInfix + "forbidden.symbols"

	FieldCreateOrDelete = "createOrDelete"
	FieldName = "name"
	FieldAliases = FieldAlias + "es"

	Create = "Create"
	Delete = "Delete"

	MaxPackageNameLen = 256
)

var noRowsWereAffected = errors.New("no rows were affected")

type PackageHandler struct {
	StateStorage wizard.StateStorage
}

func (PackageHandler) GetWizardName() string { return "PackageWizard" }
func (handler PackageHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (PackageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(packageAction)

	createOrDeleteDesc := desc.AddField(FieldCreateOrDelete, PackageFieldsTrPrefix+FieldCreateOrDelete)
	createOrDeleteDesc.InlineKeyboardAnswers = []string{Create, Delete}

	nameDesc := desc.AddField(FieldName, PackageFieldsTrPrefix+FieldName)
	nameDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxPackageNameLen {
			template := lc.Tr(PackageFieldsTrPrefix + FieldName + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return verifyNoReservedSymbols(msg.Text, lc, PackageStatusErrorForbiddenSymbolsInName)
	}

	aliasesDesc := desc.AddField(FieldAliases, PackageFieldsTrPrefix+FieldAliases)
	aliasesDesc.SkipIf = &wizard.SkipOnFieldValue{
		Name:  FieldCreateOrDelete,
		Value: Delete,
	}
	return desc
}

func (PackageHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "package" || msg.Command() == "pack"
}

func (handler PackageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	name := base.GetCommandArgument(msg)

	w := wizard.NewWizard(handler, 3)
	w.AddEmptyField(FieldCreateOrDelete, wizard.Text)
	if len(name) > 0 {
		w.AddPrefilledField(FieldName, name)
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
	}
	w.AddEmptyField(FieldAliases, wizard.Text)

	w.ProcessNextField(reqenv, msg)
}

func packageAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	deletion := fields.FindField(FieldCreateOrDelete).Data == Delete
	name := strings.ReplaceAll(fields.FindField(FieldName).Data.(string), " ", "-")

	var err error
	if deletion {
		err = deletePackage(reqenv.Ctx, reqenv.Database, uid, name)
	} else {
		aliasesStr := fields.FindField(FieldAliases).Data.(string)
		aliases := strings.Split(aliasesStr, "\n")
		aliases = funk.Map(aliases, func(a string) string {
			return strings.TrimPrefix(a, LinePrefix)
		}).([]string)

		err = createPackage(reqenv.Ctx, reqenv.Database, uid, name, aliases)
	}

	reply := replierFactory(reqenv, msg)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == DuplicateConstraintSQLCode {
		reply(PackageStatusDuplicate)
	} else if err == noRowsWereAffected {
		reply(PackageStatusNoRows)
	} else if err != nil {
		log.Error(err)
		reply(PackageStatusFailure)
	} else if deletion {
		reply(PackageStatusDeletionSuccess)
	} else {
		template := reqenv.Lang.Tr(PackageStatusCreationSuccess)
		packName := formatPackageName(uid, name)
		urlPath := url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(packName)))
		reqenv.Bot.ReplyWithMarkdown(msg, fmt.Sprintf(template, packName, packName, reqenv.Bot.GetName(), urlPath))
	}
}

func createPackage(ctx context.Context, db *sql.DB, uid int64, name string, aliases []string) error {
	var (
		tx *sql.Tx
		err error
	)
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{}); err == nil {
		if err = createPackageImpl(ctx, db, tx, uid, name, aliases); err == nil {
			err = tx.Commit()
		}
	}
	return err
}

func createPackageImpl(ctx context.Context, db *sql.DB, tx *sql.Tx, uid int64, name string, aliases []string) error {
	var (
		packID int
		res *sql.Rows
		err error
	)
	if err = tx.QueryRowContext(ctx, "INSERT INTO packages(owner_uid, name) VALUES ($1, $2) RETURNING id", uid, name).Scan(&packID); err == nil {
		aliases = funk.Map(aliases, func(a string) string {
			return strings.Replace(a, "'", "''", -1)
		}).([]string)
		if res, err = db.QueryContext(ctx, fmt.Sprintf("SELECT id FROM aliases WHERE name IN ('%s')", strings.Join(aliases, "', '"))); err == nil {
			var aliasID int
			for res.Next() {
				if err = res.Scan(&aliasID); err == nil {
					// TODO: this will be optimized in #4 (make use of batch inserts from the pgx module)
					_, err = tx.ExecContext(ctx, "INSERT INTO package_aliases(package_id, alias_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", packID, aliasID)
				}
			}
		}
	}
	return err
}

func deletePackage(ctx context.Context, db *sql.DB, uid int64, name string) error {
	res, err := db.ExecContext(ctx,"DELETE FROM packages WHERE owner_uid = $1 AND name = $2", uid, name)
	if err != nil {
		return err
	}
	if !checkRowsWereAffected(res) {
		return noRowsWereAffected
	}
	return nil
}

func formatPackageName(uid int64, name string) string {
	return fmt.Sprintf("%d@%s", uid, name)
}
