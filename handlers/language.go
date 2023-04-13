package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
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
	StateStorage wizard.StateStorage
}

func (handler LanguageHandler) GetWizardStateStorage() wizard.StateStorage {
	return handler.StateStorage
}

func (LanguageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(languageFormAction)
	f := desc.AddField(FieldLanguage, LangParamPrompt)
	f.InlineKeyboardAnswers = []string{EnFlag, RuFlag}
	return desc
}

func (LanguageHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "language" || msg.Command() == "lang"
}

func (handler LanguageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	lang := base.GetCommandArgument(msg)
	if len(lang) > 0 {
		saveLangConfig(reqenv, msg, lang)
	} else {
		w := wizard.NewWizard(handler, 1)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.ProcessNextField(reqenv, msg)
	}
}

func languageFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	saveLangConfig(reqenv, msg, fields.FindField(FieldLanguage).Data.(string))
}

func saveLangConfig(reqenv *base.RequestEnv, msg *tgbotapi.Message, language string) {
	_, err := reqenv.Database.ExecContext(reqenv.Ctx, "UPDATE users SET language = $1 WHERE uid = $2", langFlagToCode(language), msg.From.ID)
	if err != nil {
		log.Errorln(err)
		reqenv.Bot.Reply(msg, reqenv.Lang.Tr(LanguageStatusFailure))
	} else {
		reqenv.Bot.Reply(msg, reqenv.Lang.Tr(SuccessTr))
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
