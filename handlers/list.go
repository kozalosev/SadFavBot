package handlers

import (
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
)

const (
	ListStatusTrPrefix                 = "commands.list.status."
	ListStatusSuccessFavs              = ListStatusTrPrefix + StatusSuccess + ".favs"
	ListStatusSuccessPackages          = ListStatusTrPrefix + StatusSuccess + ".packages"
	ListStatusFailure                  = ListStatusTrPrefix + StatusFailure
	ListStatusNoRowsFavs               = ListStatusTrPrefix + StatusNoRows + ".favs"
	ListStatusNoRowsPackages           = ListStatusTrPrefix + StatusNoRows + ".packages"
	ListFieldAliasesOrPackagesPromptTr = "commands.list.fields.favs.or.packages"

	FieldFavsOrPackages = "favsOrPackages"
	Favs                = "Favs"
	Packages            = "Packages"

	LinePrefix = "• "
)

type ListHandler struct {
	StateStorage wizard.StateStorage
}

func (handler ListHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler ListHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(listAction)
	f := desc.AddField(FieldFavsOrPackages, ListFieldAliasesOrPackagesPromptTr)
	f.InlineKeyboardAnswers = []string{Favs, Packages}
	return desc
}

func (ListHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "list"
}

func (handler ListHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 1)
	arg := strings.ToLower(base.GetCommandArgument(msg))
	if arg == "favs" || arg == "f" || arg == "fav" {
		w.AddPrefilledField(FieldFavsOrPackages, Favs)
	} else if arg == "packages" || arg == "p" || arg == "packs" || arg == "package" || arg == "pack" {
		w.AddPrefilledField(FieldFavsOrPackages, Packages)
	} else {
		w.AddEmptyField(FieldFavsOrPackages, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func listAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	var (
		items        []string
		successTitle string
		noRowsTitle  string
		err          error
	)
	if fields.FindField(FieldFavsOrPackages).Data.(string) == Packages {
		items, err = fetchPackages(reqenv.Ctx, reqenv.Database, msg.From.ID)
		successTitle = ListStatusSuccessPackages
		noRowsTitle = ListStatusNoRowsPackages
	} else {
		items, err = fetchAliases(reqenv.Ctx, reqenv.Database, msg.From.ID)
		successTitle = ListStatusSuccessFavs
		noRowsTitle = ListStatusNoRowsFavs
	}

	replyWith := replierFactory(reqenv, msg)
	if err != nil {
		log.Errorln(err)
		replyWith(ListStatusFailure)
	} else if len(items) == 0 {
		replyWith(noRowsTitle)
	} else {
		title := reqenv.Lang.Tr(successTitle)
		reqenv.Bot.Reply(msg, title+"\n\n"+LinePrefix+strings.Join(items, "\n"+LinePrefix))
	}
}

func fetchAliases(ctx context.Context, db *sql.DB, uid int64) ([]string, error) {
	q := "SELECT a1.name, count(a1.name), null AS link FROM favs f JOIN aliases a1 ON f.alias_id = a1.id WHERE f.uid = $1 GROUP BY a1.name " +
		"UNION " +
		"SELECT a2.name, null AS count, (SELECT name FROM aliases a WHERE a.id = l.linked_alias_id) AS link FROM links l JOIN aliases a2 ON l.alias_id = a2.id WHERE l.uid = $2 " +
		"ORDER BY name"

	if rows, err := db.QueryContext(ctx, q, uid, uid); err == nil {
		var (
			aliases []string
			alias   string
			count   *int
			link    *string
		)
		for rows.Next() {
			if err = rows.Scan(&alias, &count, &link); err == nil {
				if link != nil {
					aliases = append(aliases, fmt.Sprintf("%s → %s", alias, *link))
				} else {
					aliases = append(aliases, fmt.Sprintf("%s (%d)", alias, *count))
				}
			} else {
				log.Error("Error occurred while fetching from database: ", err)
			}
		}
		return aliases, nil
	} else {
		return nil, err
	}
}

func fetchPackages(ctx context.Context, db *sql.DB, uid int64) ([]string, error) {
	q := "SELECT p.name, count(pa.alias_id) FROM packages p JOIN package_aliases pa ON p.id = pa.package_id WHERE p.owner_uid = $1 GROUP BY p.name ORDER BY p.name"

	if rows, err := db.QueryContext(ctx, q, uid); err == nil {
		var (
			packages []string
			pkg      string
			count    int
		)
		for rows.Next() {
			if err = rows.Scan(&pkg, &count); err == nil {
				packages = append(packages, fmt.Sprintf("%s (%d)", pkg, count))
			} else {
				log.Error("Error occurred while fetching from database: ", err)
			}
		}
		return funk.Map(packages, func(s string) string {
			return formatPackageName(uid, s)
		}).([]string), nil
	} else {
		return nil, err
	}
}
