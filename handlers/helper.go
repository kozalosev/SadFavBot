package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"time"
)

var commandScopePrivateChats = []base.CommandScope{base.CommandScopeAllPrivateChats}
var commandScopePrivateAndGroupChats = []base.CommandScope{base.CommandScopeAllPrivateChats, base.CommandScopeAllGroupChats, base.CommandScopeAllChatAdmins}

var markdownEscaper = strings.NewReplacer(
	"*", "\\*",
	"_", "\\_",
	"`", "\\`")

var selfDeletionDelay time.Duration

func init() {
	var delay int64 = 60 // default
	if d, err := strconv.ParseInt(os.Getenv("SELF_DELETION_DELAY_SECS"), 10, 64); err != nil {
		log.WithField(logconst.FieldFunc, "init").
			WithField(logconst.FieldConst, "selfDeletionDelay").
			Error(err)
	} else {
		delay = d
	}
	selfDeletionDelay = time.Duration(delay) * time.Second
}

func extractFavInfo(fields wizard.Fields) (string, *dto.Fav) {
	aliasField := fields.FindField(FieldAlias)
	objectField := fields.FindField(FieldObject)

	var alias string
	if a, ok := aliasField.Data.(wizard.Txt); ok {
		alias = a.Value
	} else {
		log.WithField(logconst.FieldFunc, "extractFavInfo").
			Errorf("Invalid type for alias: %T %+v", aliasField, aliasField)
		return "", nil
	}

	if objectField.Data == nil {
		return alias, &dto.Fav{}
	}

	var (
		text     wizard.Txt
		file     wizard.File
		location wizard.LocData
		ok       bool
	)
	switch objectField.Type {
	case wizard.Text:
		text, ok = objectField.Data.(wizard.Txt)
		if !ok {
			log.WithField(logconst.FieldFunc, "extractFavInfo").
				Errorf("Invalid type: string was expected but '%T %+v' is got", objectField.Data, objectField.Data)
			return "", nil
		}
	case wizard.Location:
		location, ok = objectField.Data.(wizard.LocData)
		if !ok {
			log.WithField(logconst.FieldFunc, "extractFavInfo").
				Errorf("Invalid type: LocData was expected but '%T %+v' is got", objectField.Data, objectField.Data)
			return "", nil
		}
	default:
		file, ok = objectField.Data.(wizard.File)
		if !ok {
			log.WithField(logconst.FieldFunc, "extractFavInfo").
				Errorf("Invalid type: File was expected but '%T %+v' is got", objectField.Data, objectField.Data)
			return "", nil
		}
	}

	return alias, &dto.Fav{
		Type:     objectField.Type,
		Text:     &text,
		File:     &file,
		Location: &location,
	}
}

func isGroup(chat *tgbotapi.Chat) bool {
	return chat.IsGroup() || chat.IsSuperGroup()
}

func possiblySelfDestroyingReplier(appenv *base.ApplicationEnv, reqenv *base.RequestEnv, msg *tgbotapi.Message) func(string) {
	if msg.Chat.IsPrivate() {
		return base.NewReplier(appenv, reqenv, msg)
	}

	return func(statusKey string) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, reqenv.Lang.Tr(statusKey))
		_replyDestroyingImpl(appenv.Bot, msg, reply)
	}
}

func replyPossiblySelfDestroying(appenv *base.ApplicationEnv, msg *tgbotapi.Message, text string, buttons []tgbotapi.InlineKeyboardButton) {
	if msg.Chat.IsPrivate() {
		appenv.Bot.Reply(msg, text)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	if len(buttons) > 0 {
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(buttons...),
		)
	}
	_replyDestroyingImpl(appenv.Bot, msg, reply)
}

func _replyDestroyingImpl(bot base.ExtendedBotAPI, srcMsg *tgbotapi.Message, msgToSend tgbotapi.MessageConfig) {
	msgToSend.ReplyToMessageID = srcMsg.MessageID

	if sentMsg, err := bot.Send(msgToSend); err == nil {
		go func() {
			time.Sleep(selfDeletionDelay)
			if err := bot.Request(tgbotapi.NewDeleteMessage(sentMsg.Chat.ID, sentMsg.MessageID)); err != nil {
				log.WithField(logconst.FieldFunc, "_replyDestroyingImpl").
					Error("Couldn't delete a self-destroying message", err)
			}
		}()
	} else {
		log.WithField(logconst.FieldFunc, "_replyDestroyingImpl").
			WithField(logconst.FieldCalledObject, "ExtendedBotAPI").
			WithField(logconst.FieldCalledMethod, "Send").
			Error(err)
	}
}
