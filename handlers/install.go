package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
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
	PackageItems                = "commands.install.message.package.favs"

	FieldConfirmation = "confirmation"
)

type InstallPackageHandler struct {
	StateStorage wizard.StateStorage
}

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
	pkgInfo, err := parsePackageName(name)
	if err != nil {
		log.Error(err)
		return
	}

	packageService := repo.NewPackageService(reqenv)
	if items, err := packageService.ListAliases(pkgInfo); err == nil {
		if len(items) > 0 {
			escapedItems := funk.Map(items, markdownEscaper.Replace).([]string)
			itemsMsg := fmt.Sprintf(reqenv.Lang.Tr(PackageItems), name, LinePrefix+strings.Join(escapedItems, "\n"+LinePrefix))
			reqenv.Bot.ReplyWithMarkdown(msg, itemsMsg)
		} else {
			log.Warning("Empty package: " + name)
		}
	} else {
		log.Error(err)
	}
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

	pkgInfo, err := parsePackageName(name)
	if err != nil {
		log.Error(err)
		reply(InstallStatusFailure)
		return
	}

	packageService := repo.NewPackageService(reqenv)
	if installedAliases, err := packageService.Install(uid, pkgInfo); err == repo.NoRowsWereAffected {
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
