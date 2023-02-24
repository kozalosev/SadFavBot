package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

const StartStatusFailure = "commands.start.status." + StatusFailure

type StartHandler struct {
	StateStorage wizard.StateStorage
}

func (StartHandler) GetWizardName() string                              { return "StartWizard" }
func (handler StartHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler StartHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(func(reqenv *base.RequestEnv, fields wizard.Fields) {
		languageFormAction(reqenv, fields)
		newLang := fields.FindField(FieldLanguage).Data.(string)
		reqenv.Lang = reqenv.Lang.GetContext(newLang)
		sendHelpMessage(reqenv)
	})
	f := desc.AddField(FieldLanguage, LangParamPrompt)
	f.ReplyKeyboardAnswers = []string{EnFlag, RuFlag}
	return desc
}

func (StartHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "start"
}

func (handler StartHandler) Handle(reqenv *base.RequestEnv) {
	res, err := reqenv.Database.Exec("INSERT INTO Users(uid) VALUES ($1) ON CONFLICT DO NOTHING", reqenv.Message.From.ID)
	if err != nil {
		log.Errorln(err)
		reqenv.Reply(reqenv.Lang.Tr(StartStatusFailure))
	} else if checkRowsWereAffected(res) {
		w := wizard.NewWizard(handler, 1)
		w.AddEmptyField(FieldLanguage, wizard.Text)
		w.ProcessNextField(reqenv)
	} else {
		sendHelpMessage(reqenv)
	}
}
