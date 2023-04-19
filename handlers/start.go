package handlers

import (
	"encoding/base64"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/help"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	StartStatusFailure     = "commands.start.status." + StatusFailure
	FieldInstallingPackage = "installingPackage"
)

type StartHandler struct {
	StateStorage          wizard.StateStorage
	InstallPackageHandler *InstallPackageHandler
}

func (handler StartHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler StartHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(func(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
		languageFormAction(reqenv, msg, fields)
		newLang := langFlagToCode(fields.FindField(FieldLanguage).Data.(string))
		reqenv.Lang = reqenv.Lang.GetContext(newLang)
		help.SendHelpMessage(reqenv, msg)

		if installingPackage := fields.FindField(FieldInstallingPackage).Data.(string); len(installingPackage) > 0 {
			runWizardForInstallation(reqenv, msg, &handler, installingPackage)
		}
	})
	f := desc.AddField(FieldLanguage, LangParamPrompt)
	f.InlineKeyboardAnswers = []string{EnFlag, RuFlag}
	desc.AddField(FieldInstallingPackage, "if you see this, something went wrong")
	return desc
}

func (StartHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "start"
}

func (handler StartHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	userService := repo.NewUserService(reqenv.Ctx, reqenv.Database)
	wasCreated, err := userService.Create(msg.From.ID)

	var installingPackage string
	if err == nil {
		var pkg []byte
		if pkg, err = base64.StdEncoding.DecodeString(base.GetCommandArgument(msg)); err == nil {
			installingPackage = string(pkg)
		}
	}

	if err != nil {
		log.Error(err)
		reqenv.Bot.Reply(msg, reqenv.Lang.Tr(StartStatusFailure))
	} else if wasCreated {
		w := wizard.NewWizard(handler, 2)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.AddPrefilledField(FieldInstallingPackage, installingPackage)
		w.ProcessNextField(reqenv, msg)
	} else if len(installingPackage) > 0 {
		runWizardForInstallation(reqenv, msg, &handler, installingPackage)
	} else {
		help.SendHelpMessage(reqenv, msg)
	}
}

func runWizardForInstallation(reqenv *base.RequestEnv, msg *tgbotapi.Message, handler *StartHandler, pkgName string) {
	sendCountOfAliasesInPackage(reqenv, msg, pkgName)

	w := wizard.NewWizard(handler.InstallPackageHandler, 2)
	w.AddPrefilledField(FieldName, pkgName)
	w.AddEmptyField(FieldConfirmation, wizard.Text)
	w.ProcessNextField(reqenv, msg)
}
