package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strconv"
	"strings"
)

const (
	InstallFieldsTrPrefix       = "commands.install.fields."
	InstallStatusTrPrefix       = "commands.install.status."
	InstallStatusSuccess        = InstallStatusTrPrefix + StatusSuccess
	InstallStatusSuccessNoNames = InstallStatusTrPrefix + StatusSuccess + ".no.names"
	InstallStatusFailure        = InstallStatusTrPrefix + StatusFailure
	InstallStatusNoRows         = InstallStatusTrPrefix + StatusNoRows
	PackageItems                = "commands.install.message.package.items"

	FieldConfirmation = "confirmation"
)

type InstallPackageHandler struct {
	StateStorage wizard.StateStorage
}

func (InstallPackageHandler) GetWizardName() string { return "InstallPackageWizard" }
func (handler InstallPackageHandler) GetWizardStateStorage() wizard.StateStorage {
	return handler.StateStorage
}

func (handler InstallPackageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(installPackageAction)
	desc.AddField(FieldName, InstallFieldsTrPrefix+FieldName)

	confirmDesc := desc.AddField(FieldConfirmation, InstallFieldsTrPrefix+FieldConfirmation)
	confirmDesc.InlineKeyboardAnswers = []string{Yes, No}

	return desc
}

func (handler InstallPackageHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "install"
}

func (handler InstallPackageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 2)
	name := base.GetCommandArgument(msg)
	if len(name) > 0 {
		w.AddPrefilledField(FieldName, name)
		sendCountOfAliasesInPackage(reqenv, msg, name)
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
	}
	w.AddEmptyField(FieldConfirmation, wizard.Text)
	w.ProcessNextField(reqenv, msg)
}

func sendCountOfAliasesInPackage(reqenv *base.RequestEnv, msg *tgbotapi.Message, name string) {
	if items, err := fetchAliasesInPackage(reqenv.Ctx, reqenv.Database, name); err == nil {
		if len(items) > 0 {
			itemsMsg := fmt.Sprintf(reqenv.Lang.Tr(PackageItems), name, LinePrefix+strings.Join(items, "\n"+LinePrefix))
			reqenv.Bot.ReplyWithMarkdown(msg, itemsMsg)
		} else {
			log.Warning("Empty package: " + name)
		}
	} else {
		log.Error(err)
	}
}

func fetchAliasesInPackage(ctx context.Context, db *sql.DB, name string) (items []string, err error) {
	var pkgInfo *packageInfo
	if pkgInfo, err = parsePackageName(name); err == nil {
		var res *sql.Rows
		q := "SELECT a.name FROM package_aliases pa JOIN packages p ON p.id = pa.package_id JOIN aliases a ON pa.alias_id = a.id WHERE p.owner_uid = $1 AND p.name = $2"
		if res, err = db.QueryContext(ctx, q, pkgInfo.uid, pkgInfo.name); err == nil {
			var item string
			for res.Next() {
				if err = res.Scan(&item); err == nil {
					items = append(items, item)
				}
			}
		}
	}
	return
}

func installPackageAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	if fields.FindField(FieldConfirmation).Data == Yes {
		name := fields.FindField(FieldName).Data.(string)
		installPackageWithMessageHandling(reqenv, msg, name)
	}
}

func installPackageWithMessageHandling(reqenv *base.RequestEnv, msg *tgbotapi.Message, name string) {
	uid := msg.From.ID
	reply := replierFactory(reqenv, msg)

	if installedAliases, err := installPackage(reqenv.Ctx, reqenv.Database, uid, name); err == noRowsWereAffected {
		reply(InstallStatusNoRows)
	} else if err != nil {
		log.Error(err)
		reply(InstallStatusFailure)
	} else {
		if len(installedAliases) > 0 {
			reqenv.Bot.Reply(msg, reqenv.Lang.Tr(InstallStatusSuccess)+"\n\n"+LinePrefix+strings.Join(installedAliases, "\n"+LinePrefix))
		} else {
			reply(InstallStatusSuccessNoNames)
		}
	}
}

func installPackage(ctx context.Context, db *sql.DB, uid int64, name string) ([]string, error) {
	var (
		pkgInfo      *packageInfo
		tx           *sql.Tx
		res          *sql.Rows
		err          error
		items, links []int
	)
	if pkgInfo, err = parsePackageName(name); err == nil {
		if tx, err = db.BeginTx(ctx, &sql.TxOptions{}); err == nil {
			if items, err = installItems(ctx, tx, uid, pkgInfo); err == nil {
				if links, err = installLinks(ctx, tx, uid, pkgInfo); err == nil {
					err = tx.Commit()
				}
			}
		}
	}
	aliasIDs := append(items, links...)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Error(err)
		}
		return nil, err
	} else if len(aliasIDs) == 0 {
		return nil, noRowsWereAffected
	} else {
		aliasIDs = removeDuplicates(aliasIDs)
		aliasIDsAsStr := funk.Reduce(aliasIDs[1:], func(acc string, elem int) string {
			return acc + "," + strconv.Itoa(elem)
		}, strconv.Itoa(aliasIDs[0])).(string)

		res, err = db.QueryContext(ctx, "SELECT name FROM aliases WHERE id IN ("+aliasIDsAsStr+")")

		var installedAliases []string
		if err == nil {
			var installedAlias string
			for res.Next() {
				if err = res.Scan(&installedAlias); err == nil {
					installedAliases = append(installedAliases, installedAlias)
				} else {
					log.Error(err)
				}
			}
		}
		return installedAliases, err
	}
}

func installItems(ctx context.Context, tx *sql.Tx, uid int64, pkgInfo *packageInfo) ([]int, error) {
	res, err := tx.QueryContext(ctx, "INSERT INTO items(uid, type, alias, file_id, file_unique_id, text) "+
		"SELECT cast($1 AS bigint), i.type, i.alias, i.file_id, i.file_unique_id, i.text FROM packages p "+
		"JOIN package_aliases pa ON p.id = pa.package_id "+
		"JOIN items i ON i.uid = p.owner_uid AND i.alias = pa.alias_id "+
		"WHERE p.owner_uid = $2 AND p.name = $3 "+
		"UNION "+
		"SELECT cast($1 AS bigint), i.type, i.alias, i.file_id, i.file_unique_id, i.text FROM packages p "+
		"JOIN package_aliases pa ON p.id = pa.package_id "+
		"JOIN links l ON l.uid = p.owner_uid AND l.alias_id = pa.alias_id "+
		"JOIN items i ON i.uid = p.owner_uid AND i.alias = l.linked_alias_id "+
		"WHERE p.owner_uid = $2 AND p.name = $3 "+
		"ON CONFLICT DO NOTHING "+
		"RETURNING alias", uid, pkgInfo.uid, pkgInfo.name)
	if err == nil {
		var (
			aliasID  int
			aliasIDs []int
		)
		for res.Next() {
			if err = res.Scan(&aliasID); err == nil {
				aliasIDs = append(aliasIDs, aliasID)
			} else {
				log.Error(err)
			}
		}
		return aliasIDs, nil
	} else {
		return nil, err
	}
}

func installLinks(ctx context.Context, tx *sql.Tx, uid int64, pkgInfo *packageInfo) ([]int, error) {
	res, err := tx.QueryContext(ctx, "INSERT INTO links(uid, alias_id, linked_alias_id) "+
		"SELECT $1, l.alias_id, l.linked_alias_id FROM packages p "+
		"JOIN package_aliases pa ON p.id = pa.package_id "+
		"JOIN links l ON l.uid = p.owner_uid AND l.alias_id = pa.alias_id "+
		"WHERE p.owner_uid = $2 AND p.name = $3 "+
		"ON CONFLICT DO NOTHING "+
		"RETURNING alias_id", uid, pkgInfo.uid, pkgInfo.name)
	if err == nil {
		var (
			aliasID  int
			aliasIDs []int
		)
		for res.Next() {
			if err = res.Scan(&aliasID); err == nil {
				aliasIDs = append(aliasIDs, aliasID)
			} else {
				log.Error(err)
			}
		}
		return aliasIDs, nil
	} else {
		return nil, err
	}
}

type packageInfo struct {
	uid  int64
	name string
}

func parsePackageName(s string) (*packageInfo, error) {
	arr := strings.Split(s, "@")
	if len(arr) != 2 {
		return nil, errors.New("Unexpected package name: " + s)
	}
	uid, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		return nil, err
	}
	return &packageInfo{
		uid:  uid,
		name: arr[1],
	}, nil
}

func removeDuplicates(arr []int) []int {
	type empty struct{}
	arrMap := make(map[int]empty, len(arr))
	for _, val := range arr {
		arrMap[val] = empty{}
	}
	arr = make([]int, 0, len(arrMap))
	for val := range arrMap {
		arr = append(arr, val)
	}
	return arr
}
