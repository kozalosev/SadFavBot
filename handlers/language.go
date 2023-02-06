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

func (LanguageHandler) GetWizardName() string { return "LanguageWizard" }
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

func (handler LanguageHandler) Handle(reqenv *base.RequestEnv) {
	lang := base.GetCommandArgument(reqenv.Message)
	if len(lang) > 0 {
		saveLangConfig(reqenv, lang)
	} else {
		w := wizard.NewWizard(handler, 1)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.ProcessNextField(reqenv)
	}
}

func languageFormAction(reqenv *base.RequestEnv, fields wizard.Fields) {
	saveLangConfig(reqenv, fields.FindField(FieldLanguage).Data.(string))
}

func saveLangConfig(reqenv *base.RequestEnv, language string) {
	_, err := reqenv.Database.Exec("UPDATE users SET language = $1 WHERE uid = $2", langFlagToCode(language), reqenv.Message.From.ID)
	if err != nil {
		log.Errorln(err)
		reqenv.Reply(reqenv.Lang.Tr(LanguageStatusFailure))
	} else {
		reqenv.Reply(reqenv.Lang.Tr(SuccessTr))
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
