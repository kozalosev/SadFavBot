package common

import (
	"fmt"
	"html"
	log "log/slog"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/fp"
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

// HTMLCustomizer sends the message as HTML. Unlike legacy Markdown, it allows to escape the values entered by users
// inside the entities (see FormatHTML).
var HTMLCustomizer base.MessageCustomizer = func(msgConfig *tgbotapi.MessageConfig) {
	msgConfig.ParseMode = tgbotapi.ModeHTML
}

// FormatHTML fills the HTML template with the values, escaping them. The markup belongs to the template only, so it
// can't be escaped in the formatted text.
func FormatHTML(template string, values ...string) string {
	escaped := fp.Map(values, func(value string) any { return html.EscapeString(value) })
	return fmt.Sprintf(template, escaped...)
}

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

// TextValue returns the text value of the field, or an empty string if there is no such field or its value is not text.
func TextValue(fields wizard.Fields, name string) string {
	txt, _ := wizard.GetData[wizard.Txt](fields, name)
	return txt.Value
}

func ExtractFavInfo(fields wizard.Fields) (string, *dto.Fav) {
	alias, ok := getDataOrLog[wizard.Txt](fields.FindField(FieldAlias))
	if !ok {
		return "", nil
	}

	objectField := fields.FindField(FieldObject)
	if objectField == nil || objectField.Data == nil {
		return alias.Value, &dto.Fav{}
	}

	var (
		text     wizard.Txt
		file     wizard.File
		location wizard.LocData
	)
	switch objectField.Type {
	case wizard.Text:
		if text, ok = getDataOrLog[wizard.Txt](objectField); !ok {
			return "", nil
		}
	case wizard.Location:
		if location, ok = getDataOrLog[wizard.LocData](objectField); !ok {
			return "", nil
		}
	default:
		if file, ok = getDataOrLog[wizard.File](objectField); !ok {
			return "", nil
		}
	}

	return alias.Value, &dto.Fav{
		Type:     objectField.Type,
		Text:     &text,
		File:     &file,
		Location: &location,
	}
}

func getDataOrLog[T wizard.Value](field *wizard.Field) (data T, ok bool) {
	if field != nil {
		data, ok = field.Data.(T)
	}
	if !ok {
		logInvalidFieldType[T](field)
	}
	return data, ok
}

func logInvalidFieldType[Expected wizard.Value](field *wizard.Field) {
	var data any
	if field != nil {
		data = field.Data
	}
	log.Error("Invalid type of the field value",
		logconst.FieldFunc, "ExtractFavInfo",
		"expectedType", fmt.Sprintf("%T", *new(Expected)),
		"type", fmt.Sprintf("%T", data),
		logconst.FieldValue, data)
}
