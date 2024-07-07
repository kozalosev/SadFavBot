package privacy

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/handlers"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
)

var privacyCommands = []string{"privacy"}

// noinspection GoNameStartsWithPackageName
type PrivacyHandler struct {
	base.CommandHandlerTrait

	appenv *base.ApplicationEnv
}

func NewPrivacyHandler(appenv *base.ApplicationEnv) *PrivacyHandler {
	h := &PrivacyHandler{appenv: appenv}
	h.HandlerRefForTrait = h
	return h
}

//go:embed policy.en.md
var privacyPolicyEn string

//go:embed policy.ru.md
var privacyPolicyRu string

func (*PrivacyHandler) GetCommands() []string {
	return privacyCommands
}

func (*PrivacyHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateChats
}

func (handler *PrivacyHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	var policyMessage string
	switch reqenv.Lang.GetLanguage() {
	case handlers.RuCode:
		policyMessage = privacyPolicyRu
	default:
		policyMessage = privacyPolicyEn
	}
	handler.appenv.Bot.ReplyWithMarkdown(msg, policyMessage)
}
