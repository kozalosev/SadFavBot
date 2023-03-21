package handlers

import (
	"encoding/base64"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	StartStatusFailure = "commands.start.status." + StatusFailure
	FieldInstallingPackage = "installingPackage"
)

type StartHandler struct {
	StateStorage wizard.StateStorage
	InstallPackageHandler *InstallPackageHandler
}

func (StartHandler) GetWizardName() string                              { return "StartWizard" }
func (handler StartHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler StartHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(func(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
		languageFormAction(reqenv, msg, fields)
		newLang := langFlagToCode(fields.FindField(FieldLanguage).Data.(string))
		reqenv.Lang = reqenv.Lang.GetContext(newLang)
		sendHelpMessage(reqenv, msg)

		if installingPackage := fields.FindField(FieldInstallingPackage).Data; installingPackage != nil {
			runWizardForInstallation(reqenv, msg, &handler, installingPackage.(string))
		}
	})
	f := desc.AddField(FieldLanguage, LangParamPrompt)
	f.InlineKeyboardAnswers = []string{EnFlag, RuFlag}
	return desc
}

func (StartHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "start"
}

func (handler StartHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	res, err := reqenv.Database.ExecContext(reqenv.Ctx, "INSERT INTO Users(uid) VALUES ($1) ON CONFLICT DO NOTHING", msg.From.ID)

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
	} else if checkRowsWereAffected(res) {
		w := wizard.NewWizard(handler, 2)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.AddPrefilledField(FieldInstallingPackage, installingPackage)
		w.ProcessNextField(reqenv, msg)
	} else {
		sendHelpMessage(reqenv, msg)

		if len(installingPackage) > 0 {
			runWizardForInstallation(reqenv, msg, &handler, installingPackage)
		}
	}
}

func runWizardForInstallation(reqenv *base.RequestEnv, msg *tgbotapi.Message, handler *StartHandler, pkgName string) {
	sendCountOfAliasesInPackage(reqenv, msg, pkgName)

	w := wizard.NewWizard(handler.InstallPackageHandler, 2)
	w.AddPrefilledField(FieldName, pkgName)
	w.AddEmptyField(FieldConfirmation, wizard.Text)
	w.ProcessNextField(reqenv, msg)
}
