package help

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	callbackDataPrefix = "help:"

	helpCallbackTrPrefix       = "callbacks.help."
	helpCallbackButtonTrPrefix = helpCallbackTrPrefix + "button."
	helpCallbackCaptionInline  = helpCallbackTrPrefix + "caption.inline"
	helpCallbackCurrentPage    = helpCallbackTrPrefix + "message.current.page"
)

type helpMessageKey string

const (
	startHelpMessage helpMessageKey = "start"
	favHelpKey       helpMessageKey = "fav"
	aliasHelpKey     helpMessageKey = "alias"
	inlineHelpKey    helpMessageKey = "inline"
	packageHelpKey   helpMessageKey = "package"
	linkHelpKey      helpMessageKey = "link"
	settingsHelpKey  helpMessageKey = "settings"
)

var (
	//go:embed fav.md
	favHelpMsgEn string
	//go:embed fav.ru.md
	favHelpMsgRu string

	//go:embed alias.md
	aliasHelpMsgEn string
	//go:embed alias.ru.md
	aliasHelpMsgRu string

	//go:embed inline.md
	inlineHelpMsgEn string
	//go:embed inline.ru.md
	inlineHelpMsgRu string

	//go:embed package.md
	packageHelpMsgEn string
	//go:embed package.ru.md
	packageHelpMsgRu string

	//go:embed link.md
	linkHelpMsgEn string
	//go:embed link.ru.md
	linkHelpMsgRu string

	//go:embed settings.md
	settingsHelpMsgEn string
	//go:embed settings.ru.md
	settingsHelpMsgRu string

	helpMessagesByLang map[string]map[helpMessageKey]string
)

var (
	photoExampleInline, _ = os.LookupEnv("PHOTO_INLINE_EXAMPLE")

	markupRemover = strings.NewReplacer(
		"*", "",
		"_", "",
		"\r", "")
)

func InitMessages(maxAliasLen, maxPackageNameLen int, reservedSymbols string) {
	helpMessagesByLang = map[string]map[helpMessageKey]string{
		"en": {
			favHelpKey:      favHelpMsgEn,
			aliasHelpKey:    fmt.Sprintf(aliasHelpMsgEn, maxAliasLen, reservedSymbols),
			inlineHelpKey:   inlineHelpMsgEn,
			packageHelpKey:  fmt.Sprintf(packageHelpMsgEn, maxPackageNameLen, reservedSymbols),
			linkHelpKey:     linkHelpMsgEn,
			settingsHelpKey: settingsHelpMsgEn,
		},
		"ru": {
			favHelpKey:      favHelpMsgRu,
			aliasHelpKey:    fmt.Sprintf(aliasHelpMsgRu, maxAliasLen, reservedSymbols),
			inlineHelpKey:   inlineHelpMsgRu,
			packageHelpKey:  fmt.Sprintf(packageHelpMsgRu, maxPackageNameLen, reservedSymbols),
			linkHelpKey:     linkHelpMsgRu,
			settingsHelpKey: settingsHelpMsgRu,
		},
	}
}

type CallbackHandler struct {
	appenv *base.ApplicationEnv
}

func NewCallbackHandler(appenv *base.ApplicationEnv) *CallbackHandler {
	return &CallbackHandler{appenv: appenv}
}

func (*CallbackHandler) GetCallbackPrefix() string { return callbackDataPrefix }

func (handler *CallbackHandler) Handle(reqenv *base.RequestEnv, query *tgbotapi.CallbackQuery) {
	messages, ok := helpMessagesByLang[reqenv.Lang.GetLanguage()]
	if !ok {
		messages = helpMessagesByLang["en"]
	}

	helpKey := strings.TrimPrefix(query.Data, callbackDataPrefix)
	var msg string
	if msg, ok = messages[helpMessageKey(helpKey)]; !ok {
		log.WithField(logconst.FieldHandler, "help.CallbackHandler").
			WithField(logconst.FieldMethod, "Handle").
			Error("Unexpected help key: ", helpKey)
		msg = messages[startHelpMessage]
	}

	var (
		answer tgbotapi.Chattable
		err    error
	)
	if testHelpMessagesEquality(&msg, &query.Message.Text) {
		answer = tgbotapi.NewCallback(query.ID, reqenv.Lang.Tr(helpCallbackCurrentPage))
	} else {
		a := tgbotapi.NewEditMessageTextAndMarkup(query.Message.Chat.ID, query.Message.MessageID,
			msg, buildInlineKeyboard(reqenv.Lang))
		a.ParseMode = tgbotapi.ModeMarkdown
		answer = a
	}

	if err = handler.appenv.Bot.Request(answer); err == nil {
		err = handler.sendAdditionalMessagesIfNeeded(reqenv, query.Message, &answer, helpMessageKey(helpKey))
	}
	if err != nil {
		log.WithField(logconst.FieldHandler, "help.CallbackHandler").
			WithField(logconst.FieldMethod, "Handle").
			WithField(logconst.FieldCalledObject, "BotAPI").
			WithField(logconst.FieldCalledMethod, "Request").
			Error(err)
	}
}

func (handler *CallbackHandler) sendAdditionalMessagesIfNeeded(reqenv *base.RequestEnv, originMsg *tgbotapi.Message, answer *tgbotapi.Chattable, helpKey helpMessageKey) error {
	_, wasUpdated := (*answer).(tgbotapi.EditMessageTextConfig)
	if wasUpdated && helpKey == inlineHelpKey && len(photoExampleInline) > 0 {
		media := tgbotapi.NewPhoto(originMsg.Chat.ID, tgbotapi.FileURL(photoExampleInline))
		media.Caption = reqenv.Lang.Tr(helpCallbackCaptionInline)
		media.ReplyToMessageID = originMsg.MessageID
		return handler.appenv.Bot.Request(media)
	}
	return nil
}

func buildInlineKeyboard(lc *loc.Context) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(lc.Tr(helpCallbackButtonTrPrefix+string(favHelpKey)), callbackDataPrefix+string(favHelpKey)),
			tgbotapi.NewInlineKeyboardButtonData(lc.Tr(helpCallbackButtonTrPrefix+string(aliasHelpKey)), callbackDataPrefix+string(aliasHelpKey)),
			tgbotapi.NewInlineKeyboardButtonData(lc.Tr(helpCallbackButtonTrPrefix+string(inlineHelpKey)), callbackDataPrefix+string(inlineHelpKey)),
		), tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(lc.Tr(helpCallbackButtonTrPrefix+string(packageHelpKey)), callbackDataPrefix+string(packageHelpKey)),
			tgbotapi.NewInlineKeyboardButtonData(lc.Tr(helpCallbackButtonTrPrefix+string(linkHelpKey)), callbackDataPrefix+string(linkHelpKey)),
			tgbotapi.NewInlineKeyboardButtonData(lc.Tr(helpCallbackButtonTrPrefix+string(settingsHelpKey)), callbackDataPrefix+string(settingsHelpKey)),
		))
}

// test strings equality by the first 32 character after markup removal
func testHelpMessagesEquality(m1, m2 *string) bool {
	m1Short := (*m1)[:64]
	m2Short := (*m2)[:64]

	m1Short = markupRemover.Replace(m1Short)
	m2Short = markupRemover.Replace(m2Short)

	return m1Short[:32] == m2Short[:32]
}
