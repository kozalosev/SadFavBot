package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strconv"
	"strings"
)

const (
	InstallFieldsTrPrefix         = "commands.install.fields."
	InstallStatusTrPrefix         = "commands.install.status."
	InstallStatusSuccess          = InstallStatusTrPrefix + StatusSuccess
	InstallStatusSuccessNoNames   = InstallStatusTrPrefix + StatusSuccess + ".no.names"
	InstallStatusFailure          = InstallStatusTrPrefix + StatusFailure
	InstallStatusNoRows           = InstallStatusTrPrefix + StatusNoRows
	InstallStatusLinkToExisingFav = InstallStatusTrPrefix + "link.existing.fav"
	PackageItems                  = "commands.install.message.package.favs"

	FieldConfirmation = "confirmation"
)

type InstallPackageHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	packageService *repo.PackageService
}

func NewInstallPackageHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *InstallPackageHandler {
	h := &InstallPackageHandler{
		appenv:         appenv,
		stateStorage:   stateStorage,
		packageService: repo.NewPackageService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *InstallPackageHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *InstallPackageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.installPackageAction)
	desc.AddField(FieldName, InstallFieldsTrPrefix+FieldName)

	confirmDesc := desc.AddField(FieldConfirmation, InstallFieldsTrPrefix+FieldConfirmation)
	confirmDesc.InlineKeyboardAnswers = []string{Yes, No}

	return desc
}

func (*InstallPackageHandler) GetCommands() []string {
	return installCommands
}

func (handler *InstallPackageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 2)
	name := base.GetCommandArgument(msg)
	if len(name) > 0 {
		w.AddPrefilledField(FieldName, name)
		sendCountOfAliasesInPackage(handler, reqenv, msg, name)
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
	}
	w.AddEmptyField(FieldConfirmation, wizard.Text)
	w.ProcessNextField(reqenv, msg)
}

func sendCountOfAliasesInPackage(handler *InstallPackageHandler, reqenv *base.RequestEnv, msg *tgbotapi.Message, name string) {
	pkgInfo, err := parsePackageName(name)
	if err != nil {
		log.WithField(logconst.FieldHandler, "InstallPackageHandler").
			WithField(logconst.FieldMethod, "sendCountOfAliasesInPackage").
			WithField(logconst.FieldCalledFunc, "parsePackageName").
			Error(err)
		return
	}

	if items, err := handler.packageService.ListAliases(pkgInfo); err == nil {
		if len(items) > 0 {
			escapedItems := funk.Map(items, markdownEscaper.Replace).([]string)
			itemsMsg := fmt.Sprintf(reqenv.Lang.Tr(PackageItems), name, LinePrefix+strings.Join(escapedItems, "\n"+LinePrefix))
			handler.appenv.Bot.ReplyWithMarkdown(msg, itemsMsg)
		} else {
			log.WithField(logconst.FieldHandler, "InstallPackageHandler").
				WithField(logconst.FieldMethod, "sendCountOfAliasesInPackage").
				Warning("Empty package: " + name)
		}
	} else {
		log.WithField(logconst.FieldHandler, "InstallPackageHandler").
			WithField(logconst.FieldMethod, "sendCountOfAliasesInPackage").
			WithField(logconst.FieldCalledObject, "PackageService").
			WithField(logconst.FieldCalledFunc, "ListAliases").
			Error(err)
	}
}

func (handler *InstallPackageHandler) installPackageAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	if fields.FindField(FieldConfirmation).Data == Yes {
		name := fields.FindField(FieldName).Data.(string)
		handler.installPackageWithMessageHandling(reqenv, msg, name)
	}
}

func (handler *InstallPackageHandler) installPackageWithMessageHandling(reqenv *base.RequestEnv, msg *tgbotapi.Message, name string) {
	uid := msg.From.ID
	reply := replierFactory(handler.appenv, reqenv, msg)

	pkgInfo, err := parsePackageName(name)
	if err != nil {
		log.WithField(logconst.FieldHandler, "InstallPackageHandler").
			WithField(logconst.FieldMethod, "installPackageWithMessageHandling").
			WithField(logconst.FieldCalledFunc, "parsePackageName").
			Error(err)
		reply(InstallStatusFailure)
		return
	}

	if installedAliases, err := handler.packageService.Install(uid, pkgInfo); err == repo.NoRowsWereAffected {
		reply(InstallStatusNoRows)
	} else if isAttemptToInsertLinkForExistingFav(err) {
		reply(InstallStatusLinkToExisingFav)
	} else if err != nil {
		log.WithField(logconst.FieldHandler, "InstallPackageHandler").
			WithField(logconst.FieldMethod, "installPackageWithMessageHandling").
			WithField(logconst.FieldCalledObject, "PackageService").
			WithField(logconst.FieldCalledMethod, "Install").
			Error(err)
		reply(InstallStatusFailure)
	} else {
		if len(installedAliases) > 0 {
			handler.appenv.Bot.Reply(msg, reqenv.Lang.Tr(InstallStatusSuccess)+"\n\n"+LinePrefix+strings.Join(installedAliases, "\n"+LinePrefix))
		} else {
			reply(InstallStatusSuccessNoNames)
		}
	}
}

func parsePackageName(s string) (*repo.PackageInfo, error) {
	arr := strings.Split(s, "@")
	if len(arr) != 2 {
		return nil, errors.New("Unexpected package name: " + s)
	}
	uid, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		return nil, err
	}
	return &repo.PackageInfo{
		UID:  uid,
		Name: arr[1],
	}, nil
}
