package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type StoredObject struct {
	ID     string
	Type   wizard.FieldType
	FileID string
}

type InlineHandler struct{}

func (InlineHandler) CanHandle(_ *tgbotapi.InlineQuery) bool {
	return true
}

func (InlineHandler) Handle(reqenv *base.RequestEnv) {
	objects := funk.Map(findObjects(reqenv), generateMapper(reqenv.Lang)).([]interface{})
	answer := tgbotapi.InlineConfig{
		InlineQueryID: reqenv.InlineQuery.ID,
		Results:       objects,
		IsPersonal:    true,
	}
	if _, err := reqenv.Bot.Request(answer); err != nil {
		log.Error("error while processing inline query: ", err)
	}
}

func generateMapper(lc *loc.Context) func(object *StoredObject) interface{} {
	return func(object *StoredObject) interface{} {
		switch object.Type {
		case wizard.Image:
			return tgbotapi.NewInlineQueryResultCachedPhoto(object.ID, object.FileID)
		case wizard.Sticker:
			return tgbotapi.NewInlineQueryResultCachedSticker(object.ID, object.FileID, "хуй")
		case wizard.Video:
			return tgbotapi.NewInlineQueryResultCachedVideo(object.ID, object.FileID, "хуй")
		case wizard.Audio:
			return tgbotapi.NewInlineQueryResultCachedAudio(object.ID, object.FileID)
		case wizard.Voice:
			return tgbotapi.NewInlineQueryResultCachedVoice(object.ID, object.FileID, "войсы для пидоров")
		case wizard.Gif:
			return tgbotapi.NewInlineQueryResultCachedGIF(object.ID, object.FileID)
		default:
			log.Warning("Unsupported type: ", object)
			return tgbotapi.NewInlineQueryResultArticle(object.ID, "FU", lc.Tr("inline.errors.type.invalid"))
		}
	}
}

func findObjects(reqenv *base.RequestEnv) []*StoredObject {
	rows, err := reqenv.Database.Query("SELECT id, type, file_id FROM item WHERE uid = $1 AND alias = $2",
		reqenv.InlineQuery.From.ID, reqenv.InlineQuery.Query)
	defer rows.Close()

	var result []*StoredObject
	if err != nil {
		log.Error("error occurred: ", err)
		return result
	}
	for rows.Next() {
		var row StoredObject
		err = rows.Scan(&row.ID, &row.Type, &row.FileID)
		if err != nil {
			log.Error("error while fetching from database: ", err)
			continue
		}
		result = append(result, &row)
	}
	return result
}
