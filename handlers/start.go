package handlers

import (
	"encoding/base64"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/SadFavBot/handlers/help"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	StartStatusFailure     = "commands.start.status." + StatusFailure
	FieldInstallingPackage = "installingPackage"
)

type StartEmbeddedHandlers struct {
	Language       *LanguageHandler
	InstallPackage *InstallPackageHandler
	Save           *SaveHandler
}

type StartHandler struct {
	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	embeddedHandlers StartEmbeddedHandlers

	userService    *repo.UserService
	packageService *repo.PackageService
}

func NewStartHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage, embeddedHandlers StartEmbeddedHandlers) *StartHandler {
	return &StartHandler{
		appenv:           appenv,
		stateStorage:     stateStorage,
		embeddedHandlers: embeddedHandlers,
		userService:      repo.NewUserService(appenv),
		packageService:   repo.NewPackageService(appenv),
	}
}

func (handler *StartHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *StartHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(func(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
		handler.embeddedHandlers.Language.languageFormAction(reqenv, msg, fields)
		newLang := langFlagToCode(fields.FindField(FieldLanguage).Data.(wizard.Txt).Value)
		reqenv.Lang = reqenv.Lang.GetContext(newLang)
		help.SendHelpMessage(handler.appenv, reqenv.Lang, msg)

		installingPackage := fields.FindField(FieldInstallingPackage).Data.(wizard.Txt).Value
		if len(installingPackage) > 0 {
			handler.runWizardForInstallation(reqenv, msg, installingPackage)
		}
	})
	f := desc.AddField(FieldLanguage, LangParamPrompt)
	f.InlineKeyboardAnswers = []string{EnFlag, RuFlag}
	desc.AddField(FieldInstallingPackage, "if you see this, something went wrong")
	return desc
}

func (*StartHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	return msg.Command() == "start"
}

func (handler *StartHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	if common.IsGroup(&msg.Chat) {
		return
	}

	wasCreated, err := handler.userService.Create(msg.From.ID)

	var installingPackage string
	arg := msg.CommandArguments()
	if err == nil && len(arg) > 0 {
		if strings.HasPrefix(arg, DeepLinkSavePrefix) {
			var alias []byte
			if arg == DeepLinkStartParamEmpty {
				handler.runWizardToSave(reqenv, msg, nil)
				return
			} else if alias, err = base64.RawURLEncoding.DecodeString(strings.TrimPrefix(arg, DeepLinkSavePrefix)); err == nil {
				a := string(alias)
				handler.runWizardToSave(reqenv, msg, &a)
				return
			}
		} else {
			installingPackage, err = handler.packageService.ResolveName(arg)
		}
	}

	if err != nil {
		log.WithField(logconst.FieldHandler, "StartHandler").
			WithField(logconst.FieldMethod, "Handle").
			Error(err)
		handler.appenv.Bot.Reply(msg, reqenv.Lang.Tr(StartStatusFailure))
	} else if wasCreated {
		w := wizard.NewWizard(handler, 2)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.AddPrefilledField(FieldInstallingPackage, installingPackage)
		w.ProcessNextField(reqenv, msg)
	} else if len(installingPackage) > 0 {
		handler.runWizardForInstallation(reqenv, msg, installingPackage)
	} else {
		help.SendHelpMessage(handler.appenv, reqenv.Lang, msg)
	}
}

func (handler *StartHandler) runWizardToSave(reqenv *base.RequestEnv, msg *tgbotapi.Message, alias *string) {
	w := wizard.NewWizard(handler.embeddedHandlers.Save, 2)
	if alias == nil {
		w.AddEmptyField(FieldAlias, wizard.Text)
	} else {
		w.AddPrefilledField(FieldAlias, *alias)
	}
	w.AddEmptyField(FieldObject, wizard.Auto)
	w.ProcessNextField(reqenv, msg)
}

func (handler *StartHandler) runWizardForInstallation(reqenv *base.RequestEnv, msg *tgbotapi.Message, pkgName string) {
	sendCountOfAliasesInPackage(handler.embeddedHandlers.InstallPackage, reqenv, msg, pkgName)

	w := wizard.NewWizard(handler.embeddedHandlers.InstallPackage, 2)
	w.AddPrefilledField(FieldName, pkgName)
	w.AddEmptyField(FieldConfirmation, wizard.Text)
	w.ProcessNextField(reqenv, msg)
}
