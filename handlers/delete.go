package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	errors2 "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"regexp"
	"strings"
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

var trimCountRegex = regexp.MustCompile("\\(\\d+\\)$")

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
			template := lc.Tr(DeleteFieldsTrPrefix + FieldAlias + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return nil
	}
	aliasDesc.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		aliases, err := fetchAliases(reqenv.Ctx, reqenv.Database, msg.From.ID)
		if err != nil {
			return []string{}
		} else {
			return funk.Map(aliases, func(a string) string {
				return trimCountSuffix(a)
			}).([]string)
		}
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
	itemValues, ok := extractItemValues(fields)

	replyWith := replierFactory(reqenv, msg)
	if !ok {
		replyWith(DeleteStatusFailure)
		return
	}

	var (
		res sql.Result
		err error
	)
	if deleteAll {
		res, err = deleteByAlias(reqenv.Ctx, reqenv.Database, uid, itemValues.Alias)
	} else if itemValues.Type == wizard.Text {
		res, err = deleteByText(reqenv.Ctx, reqenv.Database, uid, itemValues.Alias, itemValues.Text)
	} else {
		res, err = deleteByFileID(reqenv.Ctx, reqenv.Database, uid, itemValues.Alias, *itemValues.File)
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

func deleteByAlias(ctx context.Context, db *sql.DB, uid int64, alias string) (sql.Result, error) {
	log.Infof("Deletion of favs and/or links with uid '%d' and alias '%s'", uid, alias)
	var (
		tx       *sql.Tx
		res      sql.Result
		resUnion sqlDeleteResultUnion
		err      error
	)
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{}); err == nil {
		if res, err = tx.ExecContext(ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2)", uid, alias); err == nil {
			if err = resUnion.Add(res); err == nil {
				if res, err = tx.ExecContext(ctx, "DELETE FROM links WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2)", uid, alias); err == nil {
					if err = resUnion.Add(res); err == nil {
						err = tx.Commit()
					}
				}
			}
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = errors2.Wrap(err, rbErr.Error()) // TODO: replace with errors.Join() when migrate to Go 1.20
			}
		}
	}
	return &resUnion, err
}

func deleteByFileID(ctx context.Context, db *sql.DB, uid int64, alias string, file wizard.File) (sql.Result, error) {
	log.Infof("Deletion of fav with uid '%d', alias '%s' and file_id '%s'", uid, alias, file.UniqueID)
	return db.ExecContext(ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2) AND file_unique_id = $3", uid, alias, file.UniqueID)
}

func deleteByText(ctx context.Context, db *sql.DB, uid int64, alias, text string) (sql.Result, error) {
	log.Infof("Deletion of fav with uid '%d', alias '%s' and text '%s'", uid, alias, text)
	return db.ExecContext(ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2) AND text_id = (SELECT id FROM texts WHERE text = $3)", uid, alias, text)
}

func trimCountSuffix(s string) string {
	if indexes := trimCountRegex.FindStringIndex(s); indexes != nil {
		return strings.TrimSpace(s[:indexes[0]])
	} else {
		return s
	}
}

type sqlDeleteResultUnion struct {
	rowsAffected []int64
}

func (s *sqlDeleteResultUnion) LastInsertId() (int64, error) {
	return 0, nil
}

func (s *sqlDeleteResultUnion) RowsAffected() (int64, error) {
	return funk.SumInt64(s.rowsAffected), nil
}

func (s *sqlDeleteResultUnion) Add(res sql.Result) error {
	if rowsAffected, err := res.RowsAffected(); err == nil {
		s.rowsAffected = append(s.rowsAffected, rowsAffected)
		return nil
	} else {
		return err
	}
}
