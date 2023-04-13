package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	ModeFieldsTrPrefix = "commands.mode.fields."
	ModeStatusTrPrefix = "commands.mode.status."
	ModeStatusSuccess  = ModeStatusTrPrefix + StatusSuccess
	ModeStatusFailure  = ModeStatusTrPrefix + StatusFailure

	ModeMessageCurrentVal = "commands.mode.message.current.value"
	Enabled               = "✅"
	Disabled              = "🚫"

	FieldSubstrSearchEnabled = "substringSearchEnabled"
)

type SearchModeHandler struct {
	StateStorage wizard.StateStorage
}

func (handler SearchModeHandler) GetWizardStateStorage() wizard.StateStorage {
	return handler.StateStorage
}

func (SearchModeHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(searchModeAction)
	f := desc.AddField(FieldSubstrSearchEnabled, ModeFieldsTrPrefix+FieldSubstrSearchEnabled)
	f.InlineKeyboardAnswers = []string{Yes, No}
	return desc
}

func (SearchModeHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "mode" || msg.Command() == "mod"
}

func (handler SearchModeHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	var currVal string
	if reqenv.Options.SubstrSearchEnabled {
		currVal = Enabled
	} else {
		currVal = Disabled
	}
	reqenv.Bot.Reply(msg, reqenv.Lang.Tr(ModeMessageCurrentVal)+currVal)

	w := wizard.NewWizard(handler, 1)
	if arg := strings.ToLower(base.GetCommandArgument(msg)); arg == "true" || arg == "1" {
		w.AddPrefilledField(FieldSubstrSearchEnabled, true)
	} else if arg == "false" || arg == "0" {
		w.AddPrefilledField(FieldSubstrSearchEnabled, false)
	} else {
		w.AddEmptyField(FieldSubstrSearchEnabled, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func searchModeAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	substrSearchEnabled := fields.FindField(FieldSubstrSearchEnabled).Data == Yes
	_, err := reqenv.Database.ExecContext(reqenv.Ctx, "UPDATE Users SET substring_search = $2 WHERE uid = $1", msg.From.ID, substrSearchEnabled)

	reply := replierFactory(reqenv, msg)
	if err != nil {
		log.Error(err)
		reply(ModeStatusFailure)
	} else {
		reply(ModeStatusSuccess)
	}
}
