package handlers

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"strings"
)

type HelpHandler struct{}

//go:embed help.md
var helpMessage string

func (HelpHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "help"
}

func (HelpHandler) Handle(reqenv *base.RequestEnv) {
	substitutedHelpText := strings.Replace(helpMessage, "{{username}}", reqenv.Message.From.FirstName, 1)
	reqenv.ReplyWithMarkdown(substitutedHelpText)
}
