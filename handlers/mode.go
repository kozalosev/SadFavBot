package handlers

import (
	log "log/slog"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
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
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	userService *repo.UserService
}

func NewSearchModeHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *SearchModeHandler {
	h := &SearchModeHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		userService:  repo.NewUserService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
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

func (*SearchModeHandler) GetCommands() []string {
	return modeCommands
}

func (*SearchModeHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateChats
}

func (handler *SearchModeHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	if common.IsGroup(&msg.Chat) {
		return
	}

	var currVal string
	opts := reqenv.Options.(*dto.UserOptions)
	if opts.SubstrSearchEnabled {
		currVal = Enabled
	} else {
		currVal = Disabled
	}
	handler.appenv.Bot.Reply(msg, reqenv.Lang.Tr(ModeMessageCurrentVal)+currVal)

	w := wizard.NewWizard(handler, 1)
	if arg := strings.ToLower(msg.CommandArguments()); arg == "true" || arg == "1" {
		w.AddPrefilledText(FieldSubstrSearchEnabled, Yes)
	} else if arg == "false" || arg == "0" {
		w.AddPrefilledText(FieldSubstrSearchEnabled, No)
	} else {
		w.AddEmptyField(FieldSubstrSearchEnabled, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func (handler *SearchModeHandler) searchModeAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	substrSearchEnabled := common.TextValue(fields, FieldSubstrSearchEnabled) == Yes

	err := handler.userService.ChangeSubstringMode(msg.From.ID, substrSearchEnabled)

	reply := base.NewReplier(handler.appenv, reqenv, msg)
	if err != nil {
		log.Error("Failed to call UserService.ChangeSubstringMode",
			logconst.FieldHandler, "SearchModeHandler",
			logconst.FieldMethod, "searchModeAction",
			logconst.FieldCalledObject, "UserService",
			logconst.FieldCalledMethod, "ChangeSubstringMode",
			logconst.FieldError, err)
		reply(ModeStatusFailure)
	} else {
		reply(ModeStatusSuccess)
	}
}
