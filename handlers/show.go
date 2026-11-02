package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	log "log/slog"
	"slices"
	"strconv"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	ShowStatusTrPrefix   = "commands.show.status."
	ShowStatusUsage      = ShowStatusTrPrefix + "usage"
	ShowStatusEmpty      = ShowStatusTrPrefix + "empty"
	ShowStatusFailure    = ShowStatusTrPrefix + StatusFailure
	ShowStatusOutdated   = ShowStatusTrPrefix + "outdated"
	ShowStatusNotOwner   = ShowStatusTrPrefix + "not.owner"
	showCallbackPrefix   = "show:"
	showTokenBytesLength = 8
)

// showTypeOrder defines the order of the favs sent by the /show command: objects of the same type go together.
var showTypeOrder = []wizard.FieldType{
	wizard.Text, wizard.Image, wizard.Gif, wizard.Video, wizard.Audio, wizard.Document,
	wizard.Sticker, wizard.Voice, wizard.VideoNote, wizard.Location,
}

type ShowHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appenv     *base.ApplicationEnv
	favService *repo.FavService
	cache      ShowCache
}

func NewShowHandler(appenv *base.ApplicationEnv, cache ShowCache) *ShowHandler {
	h := &ShowHandler{
		appenv:     appenv,
		favService: repo.NewFavsService(appenv),
		cache:      cache,
	}
	h.HandlerRefForTrait = h
	return h
}

func (*ShowHandler) GetCommands() []string {
	return showCommands
}

func (handler *ShowHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	reply := base.NewReplier(handler.appenv, reqenv, msg)

	alias := strings.TrimSpace(msg.CommandArguments())
	if len(alias) == 0 {
		// only the short form of the command is supported in group chats
		if !common.IsGroup(&msg.Chat) {
			reply(ShowStatusUsage)
		}
		return
	}

	opts := reqenv.Options.(*dto.UserOptions)
	favs, err := handler.favService.Find(msg.From.ID, alias, opts.SubstrSearchEnabled, 0)
	if err != nil {
		log.Error("Failed to call FavService.Find",
			logconst.FieldHandler, "ShowHandler",
			logconst.FieldMethod, "Handle",
			logconst.FieldCalledObject, "FavService",
			logconst.FieldCalledMethod, "Find",
			logconst.FieldError, err)
		reply(ShowStatusFailure)
		return
	}
	if len(favs) == 0 {
		reply(ShowStatusEmpty)
		return
	}
	sortFavsByType(favs)

	var keyboard *tgbotapi.InlineKeyboardMarkup
	if len(favs) > 1 {
		token, err := newShowToken()
		if err == nil {
			err = handler.cache.Put(handler.appenv.Ctx, token, &ShowEntry{OwnerUID: msg.From.ID, Favs: favs})
		}
		if err != nil {
			log.Error("Failed to put the favs into the ShowCache",
				logconst.FieldHandler, "ShowHandler",
				logconst.FieldMethod, "Handle",
				logconst.FieldCalledObject, "ShowCache",
				logconst.FieldCalledMethod, "Put",
				logconst.FieldError, err)
			reply(ShowStatusFailure)
			return
		}
		kb := buildShowKeyboard(token, 0, len(favs))
		keyboard = &kb
	}

	if _, err = handler.appenv.Bot.Send(buildShowSendRequest(msg.Chat.ID, msg.MessageID, favs[0], keyboard)); err != nil {
		logconst.NewLoggerForHandler("ShowHandler").LogFailedTelegramApiRequest(err)
	}
}

type ShowCallbackHandler struct {
	appenv *base.ApplicationEnv
	cache  ShowCache
}

func NewShowCallbackHandler(appenv *base.ApplicationEnv, cache ShowCache) *ShowCallbackHandler {
	return &ShowCallbackHandler{appenv: appenv, cache: cache}
}

func (*ShowCallbackHandler) GetCallbackPrefix() string {
	return showCallbackPrefix
}

func (handler *ShowCallbackHandler) Handle(reqenv *base.RequestEnv, query *tgbotapi.CallbackQuery) {
	answerWith := newAnswerer(handler.appenv, reqenv, query)
	logger := logconst.NewLoggerForHandler("ShowCallbackHandler")

	msg := query.Message
	data := strings.Split(strings.TrimPrefix(query.Data, showCallbackPrefix), ":")
	if msg == nil || len(data) != 3 {
		logger.Use().Error("Unexpected callback query",
			logconst.FieldObject, "CallbackQuery",
			logconst.FieldQuery, query.Data)
		answerWith.err(ShowStatusFailure)
		return
	}
	token := data[0]
	newIdx, errNew := strconv.Atoi(data[1])
	curIdx, errCur := strconv.Atoi(data[2])
	if errNew != nil || errCur != nil {
		logger.Use().Error("Unexpected format of the callback data",
			logconst.FieldObject, "CallbackQuery.Data",
			logconst.FieldQuery, query.Data)
		answerWith.err(ShowStatusFailure)
		return
	}

	entry, err := handler.cache.Get(handler.appenv.Ctx, token)
	if errors.Is(err, ErrShowCacheMiss) {
		answerWith.err(ShowStatusOutdated)
		return
	} else if err != nil {
		log.Error("Failed to call ShowCache.Get",
			logconst.FieldHandler, "ShowCallbackHandler",
			logconst.FieldMethod, "Handle",
			logconst.FieldCalledObject, "ShowCache",
			logconst.FieldCalledMethod, "Get",
			logconst.FieldError, err)
		answerWith.err(ShowStatusFailure)
		return
	}
	if query.From.ID != entry.OwnerUID {
		answerWith.err(ShowStatusNotOwner)
		return
	}

	total := len(entry.Favs)
	if curIdx < 0 || curIdx >= total {
		curIdx = 0
	}
	newIdx = (newIdx%total + total) % total // wrap around; the double remainder also normalizes negative and too big values
	if newIdx == curIdx {
		answerWith.ok()
		return
	}

	keyboard := buildShowKeyboard(token, newIdx, total)
	current, next := entry.Favs[curIdx], entry.Favs[newIdx]
	if canEditInPlace(current, next) {
		err = handler.appenv.Bot.Request(buildShowEditRequest(msg.Chat.ID, msg.MessageID, next, keyboard))
	} else {
		replyTo := 0
		if msg.ReplyToMessage != nil {
			replyTo = msg.ReplyToMessage.MessageID
		}
		err = handler.appenv.Bot.Request(tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID))
		if err == nil {
			_, err = handler.appenv.Bot.Send(buildShowSendRequest(msg.Chat.ID, replyTo, next, &keyboard))
		}
	}
	if err != nil {
		logger.LogFailedTelegramApiRequest(err)
		answerWith.err(ShowStatusFailure)
		return
	}
	answerWith.ok()
}

// sortFavsByType groups the favs by their types in the order of showTypeOrder, preserving the original order inside a group.
func sortFavsByType(favs []*dto.Fav) {
	rank := func(fav *dto.Fav) int {
		if i := slices.Index(showTypeOrder, fav.Type); i >= 0 {
			return i
		}
		return len(showTypeOrder)
	}
	slices.SortStableFunc(favs, func(a, b *dto.Fav) int {
		return rank(a) - rank(b)
	})
}

// isEditable tells if a message with an object of the type can be edited to show another object of the same type.
func isEditable(t wizard.FieldType) bool {
	switch t {
	case wizard.Text, wizard.Image, wizard.Gif, wizard.Video, wizard.Audio, wizard.Document:
		return true
	default:
		return false
	}
}

// canEditInPlace is true if the message with the current object can be edited instead of being deleted and resent.
func canEditInPlace(current, next *dto.Fav) bool {
	return current.Type == next.Type && isEditable(current.Type)
}

func newShowToken() (string, error) {
	b := make([]byte, showTokenBytesLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// buildShowKeyboard builds the "◀ n/N ▶" row. The navigation is cyclic: "◀" on the first object leads to the last one
// and "▶" on the last object leads to the first one. In Go, the remainder of a negative number is negative too
// ((-1) % 3 == -1), so total is added before taking the remainder for the previous index.
func buildShowKeyboard(token string, current, total int) tgbotapi.InlineKeyboardMarkup {
	// the callback data consists of the index to switch to and the current one, which is used to choose between editing and resending
	data := func(idx int) string {
		return fmt.Sprintf("%s%s:%d:%d", showCallbackPrefix, token, idx, current)
	}
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀", data((current-1+total)%total)),
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d/%d", current+1, total), data(current)),
		tgbotapi.NewInlineKeyboardButtonData("▶", data((current+1)%total)),
	))
}

func buildShowSendRequest(chatID int64, replyTo int, fav *dto.Fav, keyboard *tgbotapi.InlineKeyboardMarkup) tgbotapi.Chattable {
	var markup any // an untyped nil must be passed if there is no keyboard
	if keyboard != nil {
		markup = *keyboard
	}
	reply := tgbotapi.ReplyParameters{MessageID: replyTo, AllowSendingWithoutReply: true}

	switch fav.Type {
	case wizard.Text:
		cfg := tgbotapi.NewMessage(chatID, fav.Text.Value)
		cfg.Entities = fav.Text.Entities
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Image:
		cfg := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fav.File.ID))
		cfg.Caption, cfg.CaptionEntities = fav.File.Caption, fav.File.Entities
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Gif:
		cfg := tgbotapi.NewAnimation(chatID, tgbotapi.FileID(fav.File.ID))
		cfg.Caption, cfg.CaptionEntities = fav.File.Caption, fav.File.Entities
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Video:
		cfg := tgbotapi.NewVideo(chatID, tgbotapi.FileID(fav.File.ID))
		cfg.Caption, cfg.CaptionEntities = fav.File.Caption, fav.File.Entities
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Audio:
		cfg := tgbotapi.NewAudio(chatID, tgbotapi.FileID(fav.File.ID))
		cfg.Caption, cfg.CaptionEntities = fav.File.Caption, fav.File.Entities
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Document:
		cfg := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fav.File.ID))
		cfg.Caption, cfg.CaptionEntities = fav.File.Caption, fav.File.Entities
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Voice:
		cfg := tgbotapi.NewVoice(chatID, tgbotapi.FileID(fav.File.ID))
		cfg.Caption, cfg.CaptionEntities = fav.File.Caption, fav.File.Entities
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Sticker:
		cfg := tgbotapi.NewSticker(chatID, tgbotapi.FileID(fav.File.ID))
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.VideoNote:
		cfg := tgbotapi.NewVideoNote(chatID, 0, tgbotapi.FileID(fav.File.ID))
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	case wizard.Location:
		cfg := tgbotapi.NewLocation(chatID, fav.Location.Latitude, fav.Location.Longitude)
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	default:
		log.Warn("Unsupported type",
			logconst.FieldFunc, "buildShowSendRequest",
			logconst.FieldObject, fmt.Sprintf("%+v", fav))
		cfg := tgbotapi.NewMessage(chatID, string(fav.Type))
		cfg.ReplyParameters, cfg.ReplyMarkup = reply, markup
		return cfg
	}
}

// buildShowEditRequest edits the message, replacing the object with another one of the same editable type.
func buildShowEditRequest(chatID int64, messageID int, fav *dto.Fav, keyboard tgbotapi.InlineKeyboardMarkup) tgbotapi.Chattable {
	if fav.Type == wizard.Text {
		cfg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, fav.Text.Value, keyboard)
		cfg.Entities = fav.Text.Entities
		return cfg
	}

	file := tgbotapi.FileID(fav.File.ID)
	var media tgbotapi.InputMedia
	switch fav.Type {
	case wizard.Image:
		m := tgbotapi.NewInputMediaPhoto(file)
		m.Caption, m.CaptionEntities = fav.File.Caption, fav.File.Entities
		media = &m
	case wizard.Gif:
		m := tgbotapi.NewInputMediaAnimation(file)
		m.Caption, m.CaptionEntities = fav.File.Caption, fav.File.Entities
		media = &m
	case wizard.Video:
		m := tgbotapi.NewInputMediaVideo(file)
		m.Caption, m.CaptionEntities = fav.File.Caption, fav.File.Entities
		media = &m
	case wizard.Audio:
		m := tgbotapi.NewInputMediaAudio(file)
		m.Caption, m.CaptionEntities = fav.File.Caption, fav.File.Entities
		media = &m
	default: // Document
		m := tgbotapi.NewInputMediaDocument(file)
		m.Caption, m.CaptionEntities = fav.File.Caption, fav.File.Entities
		media = &m
	}
	cfg := tgbotapi.NewEditMessageMedia(chatID, messageID, media)
	cfg.ReplyMarkup = &keyboard
	return cfg
}
