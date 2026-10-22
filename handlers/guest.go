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
)

const (
	GuestEmptyQueryTr  = "guest.empty.query"
	GuestEmptyResultTr = "guest.empty.result"

	guestErrorResultID = "error"
	guestEmptyResultID = "empty"
)

// GuestFavoritesHandler answers a guest query with the first fav found by the text of the message.
// A guest query can be answered only once, so there is no pagination and no choice, unlike in the inline mode.
type GuestFavoritesHandler struct {
	appenv     *base.ApplicationEnv
	favService *repo.FavService
	mention    *regexp.Regexp
}

func NewGuestFavoritesHandler(appenv *base.ApplicationEnv) *GuestFavoritesHandler {
	return &GuestFavoritesHandler{
		appenv:     appenv,
		favService: repo.NewFavsService(appenv),
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
	return generateMapper(reqenv.Lang)(favs[0])
}

// newBotMentionRegexp builds the regexp matching the mentions of the bot.
func newBotMentionRegexp(botName string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)@` + regexp.QuoteMeta(botName) + `\b`)
}

// stripBotMention removes all mentions of the bot from the text and returns the remaining part, which is an alias.
func stripBotMention(text string, mention *regexp.Regexp) string {
	return strings.TrimSpace(mention.ReplaceAllString(text, ""))
}
