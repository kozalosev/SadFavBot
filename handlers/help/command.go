package help

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/loctools/go-l10n/loc"
	"strings"
)

const ruCode = "ru"

var helpCommands = []string{"help"}

type CommandHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appenv *base.ApplicationEnv
}

func NewCommandHandler(appenv *base.ApplicationEnv) *CommandHandler {
	h := &CommandHandler{appenv: appenv}
	h.HandlerRefForTrait = h
	return h
}

//go:embed help.md
var helpMessageEn string

//go:embed help.ru.md
var helpMessageRu string

func (*CommandHandler) GetCommands() []string {
	return helpCommands
}

func (*CommandHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateAndGroupChats
}

func (handler *CommandHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	SendHelpMessage(handler.appenv, reqenv.Lang, msg)
}

func SendHelpMessage(appenv *base.ApplicationEnv, lc *loc.Context, msg *tgbotapi.Message) {
	var helpMessage *string
	switch lc.GetLanguage() {
	case ruCode:
		helpMessage = &helpMessageRu
	default:
		helpMessage = &helpMessageEn
	}
	substitutedHelpText := strings.Replace(*helpMessage, "{{username}}", msg.From.FirstName, 1)
	msgCustomizer := func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		msgConfig.ReplyMarkup = buildInlineKeyboard(lc)
	}
	common.ReplyPossiblySelfDestroying(appenv, msg, substitutedHelpText, msgCustomizer)
}
