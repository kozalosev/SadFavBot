package help

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"strings"
)

const ruCode = "ru"

type CommandHandler struct{}

//go:embed help.md
var helpMessageEn string

//go:embed help.ru.md
var helpMessageRu string

func (CommandHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "help"
}

func (CommandHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	SendHelpMessage(reqenv, msg)
}

func SendHelpMessage(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	var helpMessage *string
	switch reqenv.Lang.GetLanguage() {
	case ruCode:
		helpMessage = &helpMessageRu
	default:
		helpMessage = &helpMessageEn
	}
	substitutedHelpText := strings.Replace(*helpMessage, "{{username}}", msg.From.FirstName, 1)
	reqenv.Bot.ReplyWithMessageCustomizer(msg, substitutedHelpText, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		msgConfig.ReplyMarkup = buildInlineKeyboard(reqenv.Lang)
	})
}
