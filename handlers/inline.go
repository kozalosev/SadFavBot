package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

const (
	ErrorTitleTr  = "error"
	UnknownTypeTr = "inline.errors.type.invalid"
)

type StoredObject struct {
	ID     string
	Type   wizard.FieldType
	FileID *string
	Text   *string
}

type GetFavoritesInlineHandler struct{}

func (GetFavoritesInlineHandler) CanHandle(*tgbotapi.InlineQuery) bool {
	return true
}

func (GetFavoritesInlineHandler) Handle(reqenv *base.RequestEnv, query *tgbotapi.InlineQuery) {
	answer := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
	}
	if len(query.Query) > 0 {
		objects := funk.Map(findObjects(reqenv, query), generateMapper(reqenv.Lang)).([]interface{})
		answer.Results = objects
	}
	if err := reqenv.Bot.Request(answer); err != nil {
		log.Error("error while processing inline query: ", err)
	}
}

func generateMapper(lc *loc.Context) func(object *StoredObject) interface{} {
	caser := cases.Title(language.Make(lc.GetLanguage()))
	return func(object *StoredObject) interface{} {
		switch object.Type {
		case wizard.Text:
			return tgbotapi.NewInlineQueryResultArticle(object.ID, *object.Text, *object.Text)
		case wizard.Image:
			return tgbotapi.NewInlineQueryResultCachedPhoto(object.ID, *object.FileID)
		case wizard.Sticker:
			return tgbotapi.NewInlineQueryResultCachedSticker(object.ID, *object.FileID, caser.String(lc.Tr("sticker")))
		case wizard.Video:
			return tgbotapi.NewInlineQueryResultCachedVideo(object.ID, *object.FileID, caser.String(lc.Tr("video")))
		case wizard.Audio:
			return tgbotapi.NewInlineQueryResultCachedAudio(object.ID, *object.FileID)
		case wizard.Voice:
			return tgbotapi.NewInlineQueryResultCachedVoice(object.ID, *object.FileID, caser.String(lc.Tr("voice")))
		case wizard.Gif:
			return tgbotapi.NewInlineQueryResultCachedGIF(object.ID, *object.FileID)
		default:
			log.Warning("Unsupported type: ", object)
			return tgbotapi.NewInlineQueryResultArticle(object.ID, lc.Tr(ErrorTitleTr), lc.Tr(UnknownTypeTr))
		}
	}
}

func findObjects(reqenv *base.RequestEnv, query *tgbotapi.InlineQuery) []*StoredObject {
	escaper := strings.NewReplacer(
		"%", "\\%",
		"?", "\\?")
	userQuery := escaper.Replace(query.Query)
	if reqenv.Options.SubstrSearchEnabled {
		userQuery = "%" + userQuery + "%"
	}

	q := "SELECT min(i.id), type, file_id, t.text FROM items i " +
		"JOIN aliases a ON a.id = i.alias " +
		"LEFT JOIN texts t ON t.id = i.text " +
		"WHERE uid = $1 AND (name ILIKE $2 OR name = (SELECT ai_linked.name FROM links l " +
		"	JOIN aliases ai ON l.alias_id = ai.id " +
		"	JOIN aliases ai_linked ON l.linked_alias_id = ai_linked.id " +
		"	WHERE l.uid = $1 AND ai.name ILIKE $2)) " +
		"GROUP BY type, file_id, t.text"
	rows, err := reqenv.Database.QueryContext(reqenv.Ctx, q, query.From.ID, userQuery)

	var result []*StoredObject
	if err != nil {
		log.Error("error occurred: ", err)
		return result
	}
	for rows.Next() {
		var row StoredObject
		err = rows.Scan(&row.ID, &row.Type, &row.FileID, &row.Text)
		if err != nil {
			log.Error("Error occurred while fetching from database: ", err)
			continue
		}
		result = append(result, &row)
	}
	return result
}
