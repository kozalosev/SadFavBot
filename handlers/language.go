package handlers

import (
	log "log/slog"
	"slices"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/settings"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	LangParamPrompt       = "commands.language.fields.language"
	LanguageStatusFailure = "commands.language.status.failure"

	EnCode = "en"
	EnFlag = "🇺🇸"
	RuCode = "ru"
	RuFlag = "🇷🇺"
)

var supportedLangCodes = []string{EnFlag, EnCode, RuFlag, RuCode}

type LanguageHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	userService *repo.UserService
}

func NewLanguageHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *LanguageHandler {
	h := &LanguageHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		userService:  repo.NewUserService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
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

func (*LanguageHandler) GetCommands() []string {
	return languageCommands
}

func (*LanguageHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateChats
}

func (handler *LanguageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	if common.IsGroup(&msg.Chat) {
		return
	}

	lang := msg.CommandArguments()
	if len(lang) > 0 {
		handler.saveLangConfig(reqenv, msg, lang)
	} else {
		w := wizard.NewWizard(handler, 1)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.ProcessNextField(reqenv, msg)
	}
}

func (handler *LanguageHandler) languageFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	var lang string
	if slices.Contains(supportedLangCodes, msg.Text) {
		lang = msg.Text
	} else {
		lang = common.TextValue(fields, FieldLanguage)
	}
	handler.saveLangConfig(reqenv, msg, lang)
}

func (handler *LanguageHandler) saveLangConfig(reqenv *base.RequestEnv, msg *tgbotapi.Message, language string) {
	err := handler.userService.ChangeLanguage(msg.From.ID, settings.LangCode(langFlagToCode(language)))
	if err != nil {
		log.Error("Failed to call UserService.ChangeLanguage",
			logconst.FieldHandler, "LanguageHandler",
			logconst.FieldMethod, "saveLangConfig",
			logconst.FieldCalledObject, "UserService",
			logconst.FieldCalledMethod, "ChangeLanguage",
			logconst.FieldError, err)
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
