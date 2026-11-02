package handlers

import (
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/fp"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const guestShowCallbackPrefix = "gshow:"

// GuestShowCallbackHandler switches the objects of the message sent in response to a guest query with the "◀ n/N ▶"
// buttons, like [ShowCallbackHandler] does. The bot isn't a member of the chat, so the message can't be deleted and
// resent: it's edited by the inline_message_id, and that's why only the objects which can replace each other are
// put into the carousel (see [filterGuestCarousel]).
type GuestShowCallbackHandler struct {
	appenv *base.ApplicationEnv
	cache  ShowCache
}

func NewGuestShowCallbackHandler(appenv *base.ApplicationEnv, cache ShowCache) *GuestShowCallbackHandler {
	return &GuestShowCallbackHandler{appenv: appenv, cache: cache}
}

func (*GuestShowCallbackHandler) GetCallbackPrefix() string {
	return guestShowCallbackPrefix
}

func (handler *GuestShowCallbackHandler) Handle(reqenv *base.RequestEnv, query *tgbotapi.CallbackQuery) {
	answerWith := newAnswerer(handler.appenv, reqenv, query)
	logger := logconst.NewLoggerForHandler("GuestShowCallbackHandler")

	if query.InlineMessageID == "" {
		logger.Use().Error("InlineMessageID is empty",
			logconst.FieldObject, "query.InlineMessageID")
		answerWith.err(ShowStatusFailure)
		return
	}
	sw, ok := resolveShowSwitch(handler.appenv, handler.cache, guestShowCallbackPrefix, query, answerWith, logger)
	if !ok {
		return
	}

	keyboard := buildCarouselKeyboard(guestShowCallbackPrefix, sw.token, sw.newIdx, len(sw.entry.Favs))
	edit := tgbotapi.BaseEdit{InlineMessageID: query.InlineMessageID, ReplyMarkup: &keyboard}
	if err := handler.appenv.Bot.Request(buildFavEditRequest(edit, sw.entry.Favs[sw.newIdx])); err != nil {
		logger.LogFailedTelegramApiRequest(err)
		answerWith.err(ShowStatusFailure)
		return
	}
	answerWith.ok()
}

// filterGuestCarousel leaves only the favs the message with the first one can be edited to show: texts for a text,
// photos, GIFs, videos, audios and documents for any of them. Other types can't be edited at all, so the first fav
// stays alone.
func filterGuestCarousel(favs []*dto.Fav) []*dto.Fav {
	first := favs[0]
	if !isEditable(first.Type) {
		return favs[:1]
	}
	isText := first.Type == wizard.Text
	rest := fp.Filter(favs[1:], func(fav *dto.Fav) bool {
		return isEditable(fav.Type) && (fav.Type == wizard.Text) == isText
	})
	return append([]*dto.Fav{first}, rest...)
}

// withReplyMarkup attaches the keyboard to an inline query result of an editable type made by [generateMapper].
func withReplyMarkup(result tgbotapi.InlineQueryResult, keyboard tgbotapi.InlineKeyboardMarkup) tgbotapi.InlineQueryResult {
	switch r := result.(type) {
	case tgbotapi.InlineQueryResultArticle:
		r.ReplyMarkup = &keyboard
		return r
	case tgbotapi.InlineQueryResultCachedPhoto:
		r.ReplyMarkup = &keyboard
		return r
	case tgbotapi.InlineQueryResultCachedGIF:
		r.ReplyMarkup = &keyboard
		return r
	case tgbotapi.InlineQueryResultCachedVideo:
		r.ReplyMarkup = &keyboard
		return r
	case tgbotapi.InlineQueryResultCachedAudio:
		r.ReplyMarkup = &keyboard
		return r
	case tgbotapi.InlineQueryResultCachedDocument:
		r.ReplyMarkup = &keyboard
		return r
	default:
		return result
	}
}
