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
	ListStatusTrPrefix = "commands.list.status."
	ListStatusSuccessAliases  = ListStatusTrPrefix + StatusSuccess + ".aliases"
	ListStatusSuccessPackages  = ListStatusTrPrefix + StatusSuccess + ".packages"
	ListStatusFailure  = ListStatusTrPrefix + StatusFailure
	ListStatusNoRowsAliases   = ListStatusTrPrefix + StatusNoRows + ".aliases"
	ListStatusNoRowsPackages   = ListStatusTrPrefix + StatusNoRows + ".packages"
	ListFieldAliasesOrPackagesPromptTr = "commands.list.fields.aliases.or.packages"

	FieldAliasesOrPackages = "aliasesOrPackages"
	Aliases = "Aliases"
	Packages = "Packages"

	LinePrefix = "• "
)

type ListHandler struct{
	StateStorage wizard.StateStorage
}

func (ListHandler) GetWizardName() string { return "ListWizard" }
func (handler ListHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler ListHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(listAction)
	f := desc.AddField(FieldAliasesOrPackages, ListFieldAliasesOrPackagesPromptTr)
	f.InlineKeyboardAnswers = []string{Aliases, Packages}
	return desc
}

func (ListHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "list"
}

func (handler ListHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 1)
	arg := strings.ToLower(base.GetCommandArgument(msg))
	if  arg == "aliases" || arg == "a" || arg == "alias" {
		w.AddPrefilledField(FieldAliasesOrPackages, Aliases)
	} else if arg == "packages" || arg == "p" || arg == "packs" || arg == "package" || arg == "pack" {
		w.AddPrefilledField(FieldAliasesOrPackages, Packages)
	} else {
		w.AddEmptyField(FieldAliasesOrPackages, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func listAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	var (
		items []string
		successTitle string
		noRowsTitle string
		err error
	)
	if fields.FindField(FieldAliasesOrPackages).Data.(string) == Packages {
		items, err = fetchPackages(reqenv.Ctx, reqenv.Database, msg.From.ID)
		successTitle = ListStatusSuccessPackages
		noRowsTitle = ListStatusNoRowsPackages
	} else {
		items, err = fetchAliases(reqenv.Ctx, reqenv.Database, msg.From.ID)
		successTitle = ListStatusSuccessAliases
		noRowsTitle = ListStatusNoRowsAliases
	}

	replyWith := replierFactory(reqenv, msg)
	if err != nil {
		log.Errorln(err)
		replyWith(ListStatusFailure)
	} else if len(items) == 0 {
		replyWith(noRowsTitle)
	} else {
		title := reqenv.Lang.Tr(successTitle)
		reqenv.Bot.Reply(msg, title+"\n\n"+LinePrefix+strings.Join(items, "\n" + LinePrefix))
	}
}

func fetchAliases(ctx context.Context, db *sql.DB, uid int64) ([]string, error) {
	q := "SELECT a.name, count(a.name) FROM items i JOIN aliases a ON i.alias = a.id WHERE uid = $1 GROUP BY a.name ORDER BY a.name"
	return fetchCountableList(ctx, db, q, uid)
}

func fetchPackages(ctx context.Context, db *sql.DB, uid int64) ([]string, error) {
	q := "SELECT p.name, count(pa.alias_id) FROM packages p JOIN package_aliases pa ON p.id = pa.package_id WHERE p.owner_uid = $1 GROUP BY p.name ORDER BY p.name"
	res, err := fetchCountableList(ctx, db, q, uid)
	if err == nil {
		res = funk.Map(res, func(s string) string {
			return formatPackageName(uid, s)
		}).([]string)
	}
	return res, err
}

func fetchCountableList(ctx context.Context, db *sql.DB, query string, uid int64) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, uid)
	if err != nil {
		return nil, err
	}

	var items []string
	for rows.Next() {
		var (
			item  string
			count int
		)
		err = rows.Scan(&item, &count)
		if err != nil {
			log.Error("Error occurred while fetching from database: ", err)
			continue
		}
		items = append(items, fmt.Sprintf("%s (%d)", item, count))
	}
	return items, nil
}
