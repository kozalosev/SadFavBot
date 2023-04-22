package base

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/logconst"
	"github.com/kozalosev/SadFavBot/settings"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"golang.org/x/exp/slices"
	"strings"
)

const cmdTrTemplate = "commands.%s.description"

var (
	noOpCustomizer     MessageCustomizer = func(msgConfig *tgbotapi.MessageConfig) {}
	markdownCustomizer MessageCustomizer = func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
	}
)

// GetCommandArgument extracts a command argument from the text of a message.
// For example:
//   - "/foo bar" will result in "bar"
//   - "/foo" will result in ""
func GetCommandArgument(msg *tgbotapi.Message) string {
	return strings.TrimSpace(strings.TrimPrefix(msg.Text, "/"+msg.Command()))
}

func ConvertHandlersToCommands(handlers []MessageHandler) []CommandHandler {
	var commands []CommandHandler
	for _, h := range handlers {
		if cmd, ok := h.(CommandHandler); ok {
			commands = append(commands, cmd)
		}
	}
	return commands
}

func NewRequestEnv(langCtx *loc.Context, opts *settings.UserOptions) *RequestEnv {
	return &RequestEnv{
		Lang:    langCtx,
		Options: opts,
	}
}

func (t CommandHandlerTrait) CanHandle(msg *tgbotapi.Message) bool {
	return slices.Contains(t.HandlerRefForTrait.GetCommands(), msg.Command())
}

func NewBotAPI(api *tgbotapi.BotAPI) *BotAPI {
	return &BotAPI{internal: api}
}

func (bot *BotAPI) GetName() string {
	return bot.internal.Self.UserName
}

func (bot *BotAPI) SetCommands(locpool *loc.Pool, langCodes []string, handlers []CommandHandler) {
	for _, langCode := range langCodes {
		lc := locpool.GetContext(langCode)
		commands := funk.Map(handlers, func(h CommandHandler) tgbotapi.BotCommand {
			mainCmd := h.GetCommands()[0]
			description := lc.Tr(fmt.Sprintf(cmdTrTemplate, mainCmd))
			return tgbotapi.BotCommand{
				Command:     mainCmd,
				Description: description,
			}
		}).([]tgbotapi.BotCommand)

		req := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(tgbotapi.NewBotCommandScopeDefault(), langCode, commands...)

		logEntry := log.WithField(logconst.FieldFunc, "setCommands").
			WithField(logconst.FieldCalledObject, "BotAPI").
			WithField(logconst.FieldCalledMethod, "Request")
		if err := bot.Request(req); err != nil {
			logEntry.Error(err)
		} else {
			logEntry.Info("Commands were successfully updated!")
		}
	}
}

func (bot *BotAPI) ReplyWithMessageCustomizer(msg *tgbotapi.Message, text string, customizer MessageCustomizer) {
	if len(text) == 0 {
		log.WithField(logconst.FieldObject, "BotAPI").
			WithField(logconst.FieldMethod, "ReplyWithMessageCustomizer").
			Error("Empty reply for the message: " + msg.Text)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID
	customizer(&reply)
	if _, err := bot.internal.Send(reply); err != nil {
		log.WithField(logconst.FieldObject, "BotAPI").
			WithField(logconst.FieldMethod, "ReplyWithMessageCustomizer").
			WithField(logconst.FieldCalledObject, "internal").
			WithField(logconst.FieldCalledMethod, "Send").
			Error(err)
	}
}

func (bot *BotAPI) Reply(msg *tgbotapi.Message, text string) {
	bot.ReplyWithMessageCustomizer(msg, text, noOpCustomizer)
}

func (bot *BotAPI) ReplyWithMarkdown(msg *tgbotapi.Message, text string) {
	bot.ReplyWithMessageCustomizer(msg, text, markdownCustomizer)
}

func (bot *BotAPI) ReplyWithKeyboard(msg *tgbotapi.Message, text string, options []string) {
	buttons := funk.Map(options, func(s string) tgbotapi.KeyboardButton {
		return tgbotapi.NewKeyboardButton(s)
	}).([]tgbotapi.KeyboardButton)
	keyboard := tgbotapi.NewOneTimeReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(buttons...),
	)

	bot.ReplyWithMessageCustomizer(msg, text, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ReplyMarkup = keyboard
	})
}

func (bot *BotAPI) ReplyWithInlineKeyboard(msg *tgbotapi.Message, text string, buttons []tgbotapi.InlineKeyboardButton) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)
	bot.ReplyWithMessageCustomizer(msg, text, func(msgConfig *tgbotapi.MessageConfig) {
		msgConfig.ReplyMarkup = keyboard
	})
}

// Request is a simple wrapper around [tgbotapi.BotAPI.Request].
func (bot *BotAPI) Request(c tgbotapi.Chattable) error {
	_, err := bot.internal.Request(c)
	return err
}
