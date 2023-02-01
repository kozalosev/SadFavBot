package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	SaveFieldsTrPrefix = "commands.save.fields."
	SaveStatusTrPrefix = "commands.save.status."
)

type SaveHandler struct {
	StateStorage wizard.StateStorage
}

func (SaveHandler) GetWizardName() string              { return "SaveWizard" }
func (SaveHandler) GetWizardAction() wizard.FormAction { return saveFormAction }

func (handler SaveHandler) GetWizardStateStorage() wizard.StateStorage {
	return handler.StateStorage
}

func (SaveHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "save"
}

func (handler SaveHandler) Handle(reqenv *base.RequestEnv) {
	wizardForm := wizard.NewWizard(handler, 2)
	title := base.GetCommandArgument(reqenv.Message)
	if len(title) > 0 {
		wizardForm.AddPrefilledField(FieldAlias, title)
	} else {
		wizardForm.AddEmptyField(FieldAlias, reqenv.Lang.Tr(SaveFieldsTrPrefix+FieldAlias), wizard.Text)
	}
	wizardForm.AddEmptyField(FieldObject, reqenv.Lang.Tr(SaveFieldsTrPrefix+FieldObject), wizard.Auto)
	wizardForm.ProcessNextField(reqenv)
}

func saveFormAction(reqenv *base.RequestEnv, fields wizard.Fields) {
	name := fields.FindField(FieldAlias)
	object := fields.FindField(FieldObject)
	_, err := reqenv.Database.Exec("INSERT INTO item (uid, type, alias, file_id) VALUES ($1, $2, $3, $4)",
		reqenv.Message.From.ID, object.Type, name.Data, object.Data)
	if err != nil {
		log.Errorln(err.Error())
		reqenv.Reply(reqenv.Lang.Tr(SaveStatusTrPrefix + StatusFailure))
	} else {
		reqenv.Reply(reqenv.Lang.Tr(SaveStatusTrPrefix + StatusSuccess))
	}
}
