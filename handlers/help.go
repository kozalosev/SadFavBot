package handlers

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"strings"
)

type HelpHandler struct{}

//go:embed help.md
var helpMessageEn string

//go:embed help.ru.md
var helpMessageRu string

func (HelpHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "help"
}

func (HelpHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	sendHelpMessage(reqenv, msg)
}

func sendHelpMessage(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	var helpMessage *string
	switch reqenv.Lang.GetLanguage() {
	case RuCode:
		helpMessage = &helpMessageRu
	default:
		helpMessage = &helpMessageEn
	}
	substitutedHelpText := strings.Replace(*helpMessage, "{{username}}", msg.From.FirstName, 1)
	reqenv.Bot.ReplyWithMarkdown(msg, substitutedHelpText)
}
