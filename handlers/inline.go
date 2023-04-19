package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/wizard"
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
		log.Error(err)
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

func (*GetFavoritesInlineHandler) CanHandle(*tgbotapi.InlineQuery) bool {
	return true
}

func (handler *GetFavoritesInlineHandler) Handle(reqenv *base.RequestEnv, query *tgbotapi.InlineQuery) {
	answer := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     inlineAnswerCacheTime,
	}
	if len(query.Query) > 0 {
		if objects, err := handler.favService.Find(query.From.ID, query.Query, reqenv.Options.SubstrSearchEnabled); err == nil {
			answer.Results = funk.Map(objects, generateMapper(reqenv.Lang)).([]interface{})
		} else {
			log.Error(err)
		}
	}
	if err := handler.appenv.Bot.Request(answer); err != nil {
		log.Error("error while processing inline query: ", err)
	}
}

func generateMapper(lc *loc.Context) func(object *repo.Fav) interface{} {
	caser := cases.Title(language.Make(lc.GetLanguage()))
	return func(object *repo.Fav) interface{} {
		switch object.Type {
		case wizard.Text:
			return tgbotapi.NewInlineQueryResultArticle(object.ID, *object.Text, *object.Text)
		case wizard.Image:
			return tgbotapi.NewInlineQueryResultCachedPhoto(object.ID, object.File.ID)
		case wizard.Sticker:
			return tgbotapi.NewInlineQueryResultCachedSticker(object.ID, object.File.ID, caser.String(lc.Tr("sticker")))
		case wizard.Video:
			return tgbotapi.NewInlineQueryResultCachedVideo(object.ID, object.File.ID, caser.String(lc.Tr("video")))
		case wizard.Audio:
			return tgbotapi.NewInlineQueryResultCachedAudio(object.ID, object.File.ID)
		case wizard.Voice:
			return tgbotapi.NewInlineQueryResultCachedVoice(object.ID, object.File.ID, caser.String(lc.Tr("voice")))
		case wizard.Gif:
			return tgbotapi.NewInlineQueryResultCachedGIF(object.ID, object.File.ID)
		case wizard.Document:
			return tgbotapi.NewInlineQueryResultCachedDocument(object.ID, object.File.ID, caser.String(lc.Tr("document")))
		default:
			log.Warning("Unsupported type: ", object)
			return tgbotapi.NewInlineQueryResultArticle(object.ID, lc.Tr(ErrorTitleTr), lc.Tr(UnknownTypeTr))
		}
	}
}
