package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	ListStatusTrPrefix                 = "commands.list.status."
	ListStatusSuccessFavs              = ListStatusTrPrefix + StatusSuccess + ".favs"
	ListStatusSuccessPackages          = ListStatusTrPrefix + StatusSuccess + ".packages"
	ListStatusFailure                  = ListStatusTrPrefix + StatusFailure
	ListStatusNoRowsFavs               = ListStatusTrPrefix + StatusNoRows + ".favs"
	ListStatusNoRowsPackages           = ListStatusTrPrefix + StatusNoRows + ".packages"
	ListFieldAliasesOrPackagesPromptTr = "commands.list.fields.favs.or.packages"

	FieldFavsOrPackages = "favsOrPackages"
	FieldGrep           = "grep"
	Favs                = "Favs"
	Packages            = "Packages"

	LinePrefix = "• "

	callbackPrefix = "list-page:"
)

type ListHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	aliasService   *repo.AliasService
	packageService *repo.PackageService
}

func NewListHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *ListHandler {
	h := &ListHandler{
		appenv:         appenv,
		stateStorage:   stateStorage,
		aliasService:   repo.NewAliasService(appenv),
		packageService: repo.NewPackageService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *ListHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *ListHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.listAction)

	f := desc.AddField(FieldFavsOrPackages, ListFieldAliasesOrPackagesPromptTr)
	f.InlineKeyboardAnswers = []string{Favs, Packages}

	desc.AddField(FieldGrep, "if you see this, something went wrong")
	return desc
}

func (*ListHandler) GetCommands() []string {
	return listCommands
}

func (*ListHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateAndGroupChats
}

func (handler *ListHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 2)
	arg := strings.ToLower(msg.CommandArguments())
	args := strings.SplitN(arg, " ", 2)
	kind := args[0]
	var query string
	if len(args) == 2 {
		query = args[1]
	}
	if kind == "favs" || kind == "f" || kind == "fav" {
		w.AddPrefilledField(FieldFavsOrPackages, Favs)
	} else if kind == "packages" || kind == "p" || kind == "packs" || kind == "package" || kind == "pack" {
		w.AddPrefilledField(FieldFavsOrPackages, Packages)
	} else {
		w.AddEmptyField(FieldFavsOrPackages, wizard.Text)
	}
	w.AddPrefilledField(FieldGrep, query)

	if common.IsGroup(msg.Chat) && !w.AllRequiredFieldsFilled() {
		return
	}

	w.ProcessNextField(reqenv, msg)
}

func (handler *ListHandler) listAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	var (
		page         *repo.Page
		grep         string
		favsOrPacks  string
		successTitle string
		noRowsTitle  string
		err          error
	)
	if fields.FindField(FieldFavsOrPackages).Data.(wizard.Txt).Value == Packages {
		page, err = handler.packageService.ListWithCounts(msg.From.ID, "")
		successTitle = ListStatusSuccessPackages
		noRowsTitle = ListStatusNoRowsPackages
		favsOrPacks = Packages
	} else {
		grep = fields.FindField(FieldGrep).Data.(wizard.Txt).Value
		page, err = handler.aliasService.ListWithCounts(msg.From.ID, grep, "")
		successTitle = ListStatusSuccessFavs
		noRowsTitle = ListStatusNoRowsFavs
		favsOrPacks = Favs
	}

	replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
	if err != nil {
		log.WithField(logconst.FieldHandler, "ListHandler").
			WithField(logconst.FieldMethod, "listAction").
			WithField(logconst.FieldCalledMethod, "ListWithCounts").
			Error(err)
		replyWith(ListStatusFailure)
	} else if len(page.Items) == 0 {
		replyWith(noRowsTitle)
	} else {
		title := reqenv.Lang.Tr(successTitle)
		text := buildText(title, page)
		if page.HasNextPage {
			buttons := buildPaginationButtons(page, favsOrPacks, grep)
			common.ReplyPossiblySelfDestroying(handler.appenv, msg, text, common.SingleRowInlineKeyboardCustomizer(buttons))
		} else {
			common.ReplyPossiblySelfDestroying(handler.appenv, msg, text, base.NoOpCustomizer)
		}
	}
}

func buildText(title string, page *repo.Page) string {
	return title + "\n\n" + LinePrefix + strings.Join(page.Items, "\n"+LinePrefix)
}

func buildPaginationButtons(page *repo.Page, favsOrPacks, grep string) []tgbotapi.InlineKeyboardButton {
	lastItem := page.GetLastItem()
	lastItem = substringFromUnicodeString(lastItem, 16)
	lastItem = url.QueryEscape(lastItem)

	btnData := fmt.Sprintf("%s%s:%s:%s", callbackPrefix, favsOrPacks, grep, lastItem)
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("➡️", btnData),
	}

}

func substringFromUnicodeString(s string, limit int) string {
	var bytes []byte
	for i, v := range s {
		if i >= limit {
			break
		}
		bytes = utf8.AppendRune(bytes, v)
	}
	return string(bytes)
}

type ListPaginationCallbackHandler struct {
	appenv         *base.ApplicationEnv
	aliasService   *repo.AliasService
	packageService *repo.PackageService
}

func NewListPaginationCallbackHandler(appenv *base.ApplicationEnv) *ListPaginationCallbackHandler {
	return &ListPaginationCallbackHandler{
		appenv:         appenv,
		aliasService:   repo.NewAliasService(appenv),
		packageService: repo.NewPackageService(appenv),
	}
}

func (handler *ListPaginationCallbackHandler) GetCallbackPrefix() string {
	return callbackPrefix
}

func (handler *ListPaginationCallbackHandler) Handle(reqenv *base.RequestEnv, query *tgbotapi.CallbackQuery) {
	data := strings.Split(strings.TrimPrefix(query.Data, callbackPrefix), ":")
	answerWith := newAnswerer(handler.appenv, reqenv, query)
	logTemplate := log.
		WithField(logconst.FieldHandler, "ListPaginationCallbackHandler").
		WithField(logconst.FieldMethod, "Handle")

	msg := query.Message
	if msg == nil {
		logTemplate.WithField(logconst.FieldObject, "query.Message").
			Error("Message is nil")
		answerWith.err(ListStatusFailure)
		return
	}

	if len(data) != 3 {
		logTemplate.WithField(logconst.FieldObject, "CallbackQuery.Data").
			Error("unexpected format of the callback data: " + query.Data)
		answerWith.err(ListStatusFailure)
		return
	}

	favsOrPacks := data[0]
	grep := data[1]
	lastItem, err := url.QueryUnescape(data[2])

	if err != nil {
		logTemplate.WithField(logconst.FieldObject, "CallbackQuery.Data[2]").
			WithField(logconst.FieldCalledFunc, "QueryUnescape").
			Error(err)
		answerWith.err(ListStatusFailure)
		return
	}

	var (
		page         *repo.Page
		successTitle string
	)
	switch favsOrPacks {
	case Favs:
		page, err = handler.aliasService.ListWithCounts(query.From.ID, grep, lastItem)
		successTitle = ListStatusSuccessFavs
	case Packages:
		page, err = handler.packageService.ListWithCounts(query.From.ID, lastItem)
		successTitle = ListStatusSuccessPackages
	default:
		logTemplate.WithField(logconst.FieldObject, "favsOrPacks").
			Error("unexpected value of the favsOrPackages field: " + favsOrPacks)
		answerWith.err(ListStatusFailure)
		return
	}
	if err != nil {
		logTemplate.WithField(logconst.FieldCalledMethod, "ListWithCounts").
			Error(err)
		answerWith.err(ListStatusFailure)
		return
	}

	title := reqenv.Lang.Tr(successTitle)
	text := buildText(title, page)
	var editReq tgbotapi.EditMessageTextConfig
	if page.HasNextPage {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buildPaginationButtons(page, favsOrPacks, grep))
		editReq = tgbotapi.NewEditMessageTextAndMarkup(msg.Chat.ID, msg.MessageID, text, keyboard)
	} else {
		editReq = tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, text)
	}
	if err = handler.appenv.Bot.Request(editReq); err != nil {
		logTemplate.
			WithField(logconst.FieldCalledObject, "Bot").
			WithField(logconst.FieldCalledMethod, "Request").
			Error(err)
		answerWith.err(ListStatusFailure)
	} else {
		answerWith.ok()
	}
}

type answerer struct {
	appenv *base.ApplicationEnv
	reqenv *base.RequestEnv
	query  *tgbotapi.CallbackQuery
}

func newAnswerer(appenv *base.ApplicationEnv, reqenv *base.RequestEnv, query *tgbotapi.CallbackQuery) *answerer {
	return &answerer{
		appenv: appenv,
		reqenv: reqenv,
		query:  query,
	}
}

func (a *answerer) ok() {
	req := tgbotapi.CallbackConfig{CallbackQueryID: a.query.ID}
	if err := a.appenv.Bot.Request(req); err != nil {
		log.WithField(logconst.FieldObject, "list.answerer").
			WithField(logconst.FieldMethod, "ok").
			WithField(logconst.FieldCalledObject, "Bot").
			WithField(logconst.FieldCalledMethod, "Request").
			Error(err)
	}
}

func (a *answerer) err(key string) {
	err := a.reqenv.Lang.Tr(key)
	req := tgbotapi.NewCallbackWithAlert(a.query.ID, err)
	if err := a.appenv.Bot.Request(req); err != nil {
		log.WithField(logconst.FieldObject, "list.answerer").
			WithField(logconst.FieldMethod, "err").
			WithField(logconst.FieldCalledObject, "Bot").
			WithField(logconst.FieldCalledMethod, "Request").
			Error(err)
	}
}
