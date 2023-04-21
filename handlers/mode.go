package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/logconst"
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
	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	userService *repo.UserService
}

func NewSearchModeHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *SearchModeHandler {
	return &SearchModeHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		userService:  repo.NewUserService(appenv),
	}
}

func (handler *SearchModeHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *SearchModeHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.searchModeAction)
	f := desc.AddField(FieldSubstrSearchEnabled, ModeFieldsTrPrefix+FieldSubstrSearchEnabled)
	f.InlineKeyboardAnswers = []string{Yes, No}
	return desc
}

func (*SearchModeHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "mode" || msg.Command() == "mod"
}

func (handler *SearchModeHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	var currVal string
	if reqenv.Options.SubstrSearchEnabled {
		currVal = Enabled
	} else {
		currVal = Disabled
	}
	handler.appenv.Bot.Reply(msg, reqenv.Lang.Tr(ModeMessageCurrentVal)+currVal)

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

func (handler *SearchModeHandler) searchModeAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	substrSearchEnabled := fields.FindField(FieldSubstrSearchEnabled).Data == Yes

	err := handler.userService.ChangeSubstringMode(msg.From.ID, substrSearchEnabled)

	reply := replierFactory(handler.appenv, reqenv, msg)
	if err != nil {
		log.WithField(logconst.FieldHandler, "SearchModeHandler").
			WithField(logconst.FieldMethod, "searchModeAction").
			WithField(logconst.FieldCalledObject, "UserService").
			WithField(logconst.FieldCalledMethod, "ChangeSubstringMode").
			Error(err)
		reply(ModeStatusFailure)
	} else {
		reply(ModeStatusSuccess)
	}
}
