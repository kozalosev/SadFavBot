package help

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/loctools/go-l10n/loc"
	"strings"
)

const ruCode = "ru"

var helpCommands = []string{"help"}

type CommandHandler struct {
	base.CommandHandlerTrait

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

func (handler *CommandHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	SendHelpMessage(handler.appenv.Bot, reqenv.Lang, msg)
}

func SendHelpMessage(bot base.ExtendedBotAPI, lc *loc.Context, msg *tgbotapi.Message) {
	var helpMessage *string
	switch lc.GetLanguage() {
	case ruCode:
		helpMessage = &helpMessageRu
	default:
		helpMessage = &helpMessageEn
	}
	substitutedHelpText := strings.Replace(*helpMessage, "{{username}}", msg.From.FirstName, 1)
	bot.ReplyWithMessageCustomizer(msg, substitutedHelpText, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		msgConfig.ReplyMarkup = buildInlineKeyboard(lc)
	})
}
