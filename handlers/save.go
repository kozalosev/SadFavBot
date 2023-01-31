package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"strings"
)

type SaveHandler struct {
	StateStorage wizard.StateStorage
}

func (SaveHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "save"
}

func (handler SaveHandler) Handle(reqenv *base.RequestEnv) {
	wizardForm := wizard.NewWizard(handler, 2, handler.StateStorage)
	title := strings.TrimSpace(strings.TrimPrefix(reqenv.Message.Text, "/"+reqenv.Message.Command()))
	if len(title) > 0 {
		wizardForm.AddPrefilledField("name", title)
	} else {
		wizardForm.AddEmptyField("name", reqenv.Lang.Tr("commands.save.fields.name"), wizard.Text)
	}
	wizardForm.AddEmptyField("object", reqenv.Lang.Tr("commands.save.fields.object"), wizard.Auto)
	wizardForm.ProcessNextField(reqenv)
}

func (handler SaveHandler) GetWizardName() string {
	return "SaveWizard"
}

func (handler SaveHandler) GetWizardAction() wizard.FormAction {
	return saveFormAction
}

func saveFormAction(reqenv *base.RequestEnv, fields wizard.Fields) {
	name := fields.FindField("name")
	object := fields.FindField("object")
	_, err := reqenv.Database.Exec("INSERT INTO item (uid, type, alias, file_id) VALUES ($1, $2, $3, $4)",
		reqenv.Message.From.ID, object.Type, name.Data, object.Data)
	if err != nil {
		reqenv.Reply(err.Error())
	} else {
		reqenv.Reply(reqenv.Lang.Tr("commands.save.status.success"))
	}
}
