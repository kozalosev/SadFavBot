package handlers

import (
	"fmt"
	log "log/slog"
	"regexp"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	GuestEmptyQueryTr  = "guest.empty.query"
	GuestEmptyResultTr = "guest.empty.result"

	guestErrorResultID = "error"
	guestEmptyResultID = "empty"
)

// guestCarouselCounter counts the answers to guest queries sent with the buttons to switch between favs.
var guestCarouselCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "guest_carousel_was_shown",
	Help: "Usage counter",
})

// GuestFavoritesHandler answers a guest query with the first fav found by the text of the message.
// A guest query can be answered only once, so there is no choice, unlike in the inline mode. Instead, the other favs
// the message can be edited to show are available via the "◀ n/N ▶" buttons handled by [GuestShowCallbackHandler].
type GuestFavoritesHandler struct {
	appenv     *base.ApplicationEnv
	favService *repo.FavService
	cache      ShowCache
	mention    *regexp.Regexp
}

func NewGuestFavoritesHandler(appenv *base.ApplicationEnv, cache ShowCache) *GuestFavoritesHandler {
	return &GuestFavoritesHandler{
		appenv:     appenv,
		favService: repo.NewFavsService(appenv),
		cache:      cache,
		mention:    newBotMentionRegexp(appenv.Bot.GetName()),
	}
}

func (*GuestFavoritesHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	return msg.GuestQueryID != ""
}

func (handler *GuestFavoritesHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) tgbotapi.InlineQueryResult {
	user := msg.From
	if user == nil {
		user = msg.GuestBotCallerUser
	}
	if user == nil {
		return nil
	}

	query := stripBotMention(msg.Text, handler.mention)
	if len(query) == 0 {
		text := reqenv.Lang.Tr(GuestEmptyQueryTr)
		return tgbotapi.NewInlineQueryResultArticle(guestEmptyResultID, text, text)
	}

	opts := reqenv.Options.(*dto.UserOptions)
	favs, err := handler.favService.Find(user.ID, query, opts.SubstrSearchEnabled, 0)
	if err != nil {
		log.Error("Failed to call FavService.Find",
			logconst.FieldHandler, "GuestFavoritesHandler",
			logconst.FieldMethod, "Handle",
			logconst.FieldCalledObject, "FavService",
			logconst.FieldCalledMethod, "Find",
			logconst.FieldError, err)
		return tgbotapi.NewInlineQueryResultArticle(guestErrorResultID, reqenv.Lang.Tr(ErrorTitleTr), reqenv.Lang.Tr(ErrorTitleTr))
	}
	if len(favs) == 0 {
		text := fmt.Sprintf(reqenv.Lang.Tr(GuestEmptyResultTr), query)
		return tgbotapi.NewInlineQueryResultArticle(guestEmptyResultID, text, text)
	}

	result := generateMapper(reqenv.Lang)(favs[0])
	carousel := filterGuestCarousel(favs)
	if len(carousel) < 2 {
		return result
	}
	token, err := newShowToken()
	if err == nil {
		err = handler.cache.Put(handler.appenv.Ctx, token, &ShowEntry{OwnerUID: user.ID, Favs: carousel})
	}
	if err != nil {
		// the fav can still be sent, just without the buttons
		log.Error("Failed to put the favs into the ShowCache",
			logconst.FieldHandler, "GuestFavoritesHandler",
			logconst.FieldMethod, "Handle",
			logconst.FieldCalledObject, "ShowCache",
			logconst.FieldCalledMethod, "Put",
			logconst.FieldError, err)
		return result
	}
	guestCarouselCounter.Inc()
	return withReplyMarkup(result, buildCarouselKeyboard(guestShowCallbackPrefix, token, 0, len(carousel)))
}

// newBotMentionRegexp builds the regexp matching the mentions of the bot
// along with the commas of the direct address ("alias, @bot" or "@bot, alias").
func newBotMentionRegexp(botName string) *regexp.Regexp {
	return regexp.MustCompile(`(?i),?\s*@` + regexp.QuoteMeta(botName) + `\b,?`)
}

// stripBotMention removes all mentions of the bot from the text and returns the remaining part, which is an alias.
func stripBotMention(text string, mention *regexp.Regexp) string {
	return strings.TrimSpace(mention.ReplaceAllString(text, ""))
}
