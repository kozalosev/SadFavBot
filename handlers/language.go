package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	LangParamPrompt       = "commands.language.fields.language"
	LanguageStatusFailure = "commands.language.status.failure"

	EnCode = "en"
	EnFlag = "🇺🇸"
	RuCode = "ru"
	RuFlag = "🇷🇺"
)

type LanguageHandler struct {
	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	userService *repo.UserService
}

func NewLanguageHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *LanguageHandler {
	return &LanguageHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		userService:  repo.NewUserService(appenv),
	}
}

func (handler *LanguageHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *LanguageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.languageFormAction)
	f := desc.AddField(FieldLanguage, LangParamPrompt)
	f.InlineKeyboardAnswers = []string{EnFlag, RuFlag}
	return desc
}

func (*LanguageHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "language" || msg.Command() == "lang"
}

func (handler *LanguageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	lang := base.GetCommandArgument(msg)
	if len(lang) > 0 {
		handler.saveLangConfig(reqenv, msg, lang)
	} else {
		w := wizard.NewWizard(handler, 1)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.ProcessNextField(reqenv, msg)
	}
}

func (handler *LanguageHandler) languageFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	handler.saveLangConfig(reqenv, msg, fields.FindField(FieldLanguage).Data.(string))
}

func (handler *LanguageHandler) saveLangConfig(reqenv *base.RequestEnv, msg *tgbotapi.Message, language string) {
	err := handler.userService.ChangeLanguage(msg.From.ID, settings.LangCode(langFlagToCode(language)))
	if err != nil {
		log.Errorln(err)
		handler.appenv.Bot.Reply(msg, reqenv.Lang.Tr(LanguageStatusFailure))
	} else {
		handler.appenv.Bot.Reply(msg, reqenv.Lang.Tr(SuccessTr))
	}
}

func langFlagToCode(flag string) string {
	switch flag {
	case EnFlag:
		return EnCode
	case RuFlag:
		return RuCode
	default:
		return flag
	}
}
