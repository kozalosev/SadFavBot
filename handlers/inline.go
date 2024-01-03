package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"strconv"
)

const (
	ErrorTitleTr  = "error"
	UnknownTypeTr = "inline.errors.type.invalid"
)

var inlineAnswerCacheTime int

func init() {
	if cacheTime, err := strconv.Atoi(os.Getenv("INLINE_CACHE_TIME")); err != nil {
		log.WithField(logconst.FieldFunc, "init").
			WithField(logconst.FieldConst, "INLINE_CACHE_TIME").
			Error(err)
		inlineAnswerCacheTime = 300 // default
	} else {
		inlineAnswerCacheTime = cacheTime
	}
}

type GetFavoritesInlineHandler struct {
	appenv     *base.ApplicationEnv
	favService *repo.FavService
}

func NewGetFavoritesInlineHandler(appenv *base.ApplicationEnv) *GetFavoritesInlineHandler {
	return &GetFavoritesInlineHandler{
		appenv:     appenv,
		favService: repo.NewFavsService(appenv),
	}
}

func (*GetFavoritesInlineHandler) CanHandle(*base.RequestEnv, *tgbotapi.InlineQuery) bool {
	return true
}

func (handler *GetFavoritesInlineHandler) Handle(reqenv *base.RequestEnv, query *tgbotapi.InlineQuery) {
	answer := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     inlineAnswerCacheTime,
	}
	if len(query.Query) > 0 {
		opts := reqenv.Options.(*dto.UserOptions)
		if objects, err := handler.favService.Find(query.From.ID, query.Query, opts.SubstrSearchEnabled); err == nil {
			answer.Results = funk.Map(objects, generateMapper(reqenv.Lang)).([]interface{})
		} else {
			log.WithField(logconst.FieldHandler, "GetFavoritesInlineHandler").
				WithField(logconst.FieldMethod, "Handle").
				WithField(logconst.FieldCalledObject, "FavService").
				WithField(logconst.FieldCalledMethod, "Find").
				Error(err)
		}
	}
	if err := handler.appenv.Bot.Request(answer); err != nil {
		log.WithField(logconst.FieldHandler, "GetFavoritesInlineHandler").
			WithField(logconst.FieldMethod, "Handle").
			WithField(logconst.FieldCalledObject, "BotAPI").
			WithField(logconst.FieldCalledMethod, "Request").
			Error("Telegram Bot API request error: ", err)
	}
}

func generateMapper(lc *loc.Context) func(object *dto.Fav) interface{} {
	caser := cases.Title(language.Make(lc.GetLanguage()))
	return func(object *dto.Fav) interface{} {
		switch object.Type {
		case wizard.Text:
			article := tgbotapi.NewInlineQueryResultArticle(object.ID, object.Text.Value, object.Text.Value)

			// this is a copy, not a reference to the original InputTextMessageContent!
			content := article.InputMessageContent.(tgbotapi.InputTextMessageContent)
			content.Entities = object.Text.Entities

			article.InputMessageContent = content
			return article
		case wizard.Image:
			photo := tgbotapi.NewInlineQueryResultCachedPhoto(object.ID, object.File.ID)
			photo.Caption = object.File.Caption
			photo.CaptionEntities = object.File.Entities
			return photo
		case wizard.Sticker:
			return tgbotapi.NewInlineQueryResultCachedSticker(object.ID, object.File.ID, caser.String(lc.Tr("sticker")))
		case wizard.Video:
			video := tgbotapi.NewInlineQueryResultCachedVideo(object.ID, object.File.ID, caser.String(lc.Tr("video")))
			video.Caption = object.File.Caption
			video.CaptionEntities = object.File.Entities
			return video
		case wizard.Audio:
			audio := tgbotapi.NewInlineQueryResultCachedAudio(object.ID, object.File.ID)
			audio.Caption = object.File.Caption
			audio.CaptionEntities = object.File.Entities
			return audio
		case wizard.Voice:
			voice := tgbotapi.NewInlineQueryResultCachedVoice(object.ID, object.File.ID, caser.String(lc.Tr("voice")))
			voice.Caption = object.File.Caption
			voice.CaptionEntities = object.File.Entities
			return voice
		case wizard.Gif:
			gif := tgbotapi.NewInlineQueryResultCachedGIF(object.ID, object.File.ID)
			gif.Caption = object.File.Caption
			gif.CaptionEntities = object.File.Entities
			return gif
		case wizard.Document:
			document := tgbotapi.NewInlineQueryResultCachedDocument(object.ID, object.File.ID, caser.String(lc.Tr("document")))
			document.Caption = object.File.Caption
			document.CaptionEntities = object.File.Entities
			return document
		case wizard.Location:
			return tgbotapi.NewInlineQueryResultLocation(object.ID, caser.String(lc.Tr("location")), object.Location.Latitude, object.Location.Longitude)
		default:
			log.WithField(logconst.FieldFunc, "generateMapper").
				Warning("Unsupported type: ", object)
			return tgbotapi.NewInlineQueryResultArticle(object.ID, lc.Tr(ErrorTitleTr), lc.Tr(UnknownTypeTr))
		}
	}
}
