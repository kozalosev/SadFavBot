package common

import (
	"fmt"
	log "log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	FieldAlias  = "alias"
	FieldObject = "object"
)

var CommandScopePrivateChats = []base.CommandScope{base.CommandScopeAllPrivateChats}
var CommandScopePrivateAndGroupChats = []base.CommandScope{base.CommandScopeAllPrivateChats, base.CommandScopeAllGroupChats, base.CommandScopeAllChatAdmins}

var MarkdownEscaper = strings.NewReplacer(
	"*", "\\*",
	"_", "\\_",
	"`", "\\`")

var selfDeletionDelay time.Duration

func init() {
	var delay int64 = 60 // default
	if d, err := strconv.ParseInt(os.Getenv("SELF_DELETION_DELAY_SECS"), 10, 64); err != nil {
		log.Error("Failed to parse an environment variable",
			logconst.FieldFunc, "init",
			logconst.FieldConst, "selfDeletionDelay",
			logconst.FieldError, err)
	} else {
		delay = d
	}
	selfDeletionDelay = time.Duration(delay) * time.Second
}

func ExtractFavInfo(fields wizard.Fields) (string, *dto.Fav) {
	aliasField := fields.FindField(FieldAlias)
	objectField := fields.FindField(FieldObject)

	var alias string
	if a, ok := aliasField.Data.(wizard.Txt); ok {
		alias = a.Value
	} else {
		log.Error("Invalid type for alias",
			logconst.FieldFunc, "extractFavInfo",
			"type", fmt.Sprintf("%T", aliasField),
			"value", fmt.Sprintf("%+v", aliasField))
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
			log.Error("Invalid type: string was expected",
				logconst.FieldFunc, "extractFavInfo",
				"type", fmt.Sprintf("%T", objectField.Data),
				"value", fmt.Sprintf("%+v", objectField.Data))
			return "", nil
		}
	case wizard.Location:
		location, ok = objectField.Data.(wizard.LocData)
		if !ok {
			log.Error("Invalid type: LocData was expected",
				logconst.FieldFunc, "extractFavInfo",
				"type", fmt.Sprintf("%T", objectField.Data),
				"value", fmt.Sprintf("%+v", objectField.Data))
			return "", nil
		}
	default:
		file, ok = objectField.Data.(wizard.File)
		if !ok {
			log.Error("Invalid type: File was expected",
				logconst.FieldFunc, "extractFavInfo",
				"type", fmt.Sprintf("%T", objectField.Data),
				"value", fmt.Sprintf("%+v", objectField.Data))
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

func IsGroup(chat *tgbotapi.Chat) bool {
	return chat.IsGroup() || chat.IsSuperGroup()
}
