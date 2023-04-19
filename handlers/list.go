package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
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
		packageService := repo.NewPackageService(reqenv)
		items, err = packageService.ListWithCounts(msg.From.ID)
		successTitle = ListStatusSuccessPackages
		noRowsTitle = ListStatusNoRowsPackages
	} else {
		aliasService := repo.NewAliasService(reqenv)
		items, err = aliasService.ListWithCounts(msg.From.ID)
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
