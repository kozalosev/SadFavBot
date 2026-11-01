package handlers

import (
	"fmt"
	log "log/slog"
	"regexp"
	"strings"
	"unicode"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

const (
	GuestHintReplyTr       = "guest.hint.reply"
	GuestDeleteHintAliasTr = "guest.delete.hint.alias"
	GuestLinkHintTr        = "guest.link.hint"
	GuestRmLinkHintTr      = "guest.rmlink.hint"
	GuestHelpTr            = "guest.help"
	GuestListPlaceholderTr = "guest.list.placeholder"

	guestAnswerResultID = "answer"
)

type guestCommand int

const (
	guestCmdSave guestCommand = iota + 1
	guestCmdDelete
	guestCmdRef
	guestCmdList
	guestCmdLink
	guestCmdRmLink
	guestCmdHelp
)

// guestKeywords are the words (one or two) after the mention of the bot that switch the guest mode from the search
// to a command. They must be in lowercase.
var guestKeywords = map[string]guestCommand{
	"save":      guestCmdSave,
	"create":    guestCmdSave,
	"сохранить": guestCmdSave,
	"создать":   guestCmdSave,

	"delete":  guestCmdDelete,
	"remove":  guestCmdDelete,
	"удалить": guestCmdDelete,
	"удали":   guestCmdDelete,

	"ref":        guestCmdRef,
	"reference":  guestCmdRef,
	"info":       guestCmdRef,
	"реф":        guestCmdRef,
	"инфо":       guestCmdRef,
	"информация": guestCmdRef,

	"list":       guestCmdList,
	"список":     guestCmdList,
	"покажи все": guestCmdList,

	"link":      guestCmdLink,
	"ссылка":    guestCmdLink,
	"алиас":     guestCmdLink,
	"псевдоним": guestCmdLink,

	"rmlink":            guestCmdRmLink,
	"remove link":       guestCmdRmLink,
	"delete link":       guestCmdRmLink,
	"удалить ссылку":    guestCmdRmLink,
	"удалить алиас":     guestCmdRmLink,
	"удалить псевдоним": guestCmdRmLink,

	"help":    guestCmdHelp,
	"помощь":  guestCmdHelp,
	"справка": guestCmdHelp,
	"хелп":    guestCmdHelp,
}

const guestMaxKeywordWords = 2

// guestLinkSeparators separate the name of a link from the alias it refers to: "link foo -> bar".
var guestLinkSeparators = []string{"->", "→"}

// GuestCommandsHandler executes the commands written after the mention of the bot in the guest mode:
// "@bot save alias", "@bot delete alias", "@bot ref", "@bot list", "@bot link name -> alias", etc.
// The commands that work with an object take it from the message the guest query was sent in reply to.
// It must be registered before [GuestFavoritesHandler], which handles everything else.
type GuestCommandsHandler struct {
	appenv *base.ApplicationEnv

	favService     *repo.FavService
	userService    *repo.UserService
	aliasService   *repo.AliasService
	packageService *repo.PackageService
	linkService    *repo.LinkService
	mention        *regexp.Regexp
}

func NewGuestCommandsHandler(appenv *base.ApplicationEnv) *GuestCommandsHandler {
	return &GuestCommandsHandler{
		appenv:         appenv,
		favService:     repo.NewFavsService(appenv),
		userService:    repo.NewUserService(appenv),
		aliasService:   repo.NewAliasService(appenv),
		packageService: repo.NewPackageService(appenv),
		linkService:    repo.NewLinkService(appenv),
		mention:        newBotMentionRegexp(appenv.Bot.GetName()),
	}
}

func (handler *GuestCommandsHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	if msg.GuestQueryID == "" {
		return false
	}
	cmd, _, _ := parseGuestCommand(stripBotMention(msg.Text, handler.mention))
	return cmd != 0
}

func (handler *GuestCommandsHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) tgbotapi.InlineQueryResult {
	user := msg.From
	if user == nil {
		user = msg.GuestBotCallerUser
	}
	if user == nil {
		return nil
	}

	cmd, lang, args := parseGuestCommand(stripBotMention(msg.Text, handler.mention))
	// the language of the keyword is more reliable than the settings of the user
	lc := reqenv.Lang.GetContext(lang)
	result := handler.execute(cmd, lc, user.ID, msg.ReplyToMessage, args)
	if result == nil {
		return nil
	}

	// The bot isn't a member of the chat, so it can't delete its answer. The list is the only one worth hiding,
	// the others are short and help to promote the bot.
	if cmd == guestCmdList {
		return base.PlaceholderGuestResult{
			Result:      result,
			After:       common.SelfDeletionDelay(),
			Placeholder: lc.Tr(GuestListPlaceholderTr),
		}
	}
	return result
}

func (handler *GuestCommandsHandler) execute(cmd guestCommand, lc *loc.Context, uid int64, replyTo *tgbotapi.Message, args string) tgbotapi.InlineQueryResult {
	switch cmd {
	case guestCmdSave:
		return handler.save(lc, uid, replyTo, args)
	case guestCmdDelete:
		return handler.delete(lc, uid, replyTo, args)
	case guestCmdRef:
		return handler.ref(lc, uid, replyTo, args)
	case guestCmdList:
		return handler.list(lc, uid, args)
	case guestCmdLink:
		return handler.link(lc, uid, args)
	case guestCmdRmLink:
		return handler.rmLink(lc, uid, args)
	case guestCmdHelp:
		return guestMarkdownAnswer(lc.Tr(GuestHelpTr))
	default:
		return nil
	}
}

func (handler *GuestCommandsHandler) delete(lc *loc.Context, uid int64, replyTo *tgbotapi.Message, alias string) tgbotapi.InlineQueryResult {
	if problem := checkReplyArgs(lc, alias, replyTo, GuestDeleteHintAliasTr, GuestHintReplyTr); problem != "" {
		return guestMarkdownAnswer(problem)
	}
	if problem := checkName(lc, alias, DeleteFieldsTrPrefix+FieldAlias+FieldMaxLengthErrorTrSuffix, ""); problem != "" {
		return guestAnswer(problem)
	}
	fav := favFromMessage(replyTo)
	if fav == nil {
		return guestAnswer(lc.Tr(DeleteStatusNoRows))
	}

	res, err := handler.favService.DeleteFav(uid, alias, fav)
	if err != nil {
		return handler.failure(err, "FavService", "DeleteFav", lc.Tr(DeleteStatusFailure))
	}
	if res.RowsAffected() == 0 {
		return guestAnswer(lc.Tr(DeleteStatusNoRows))
	}
	return guestAnswer(lc.Tr(DeleteStatusSuccess))
}

// ref shows the aliases of the object from the replied message or, if it's absent, the packages containing the alias.
func (handler *GuestCommandsHandler) ref(lc *loc.Context, uid int64, replyTo *tgbotapi.Message, alias string) tgbotapi.InlineQueryResult {
	if replyTo == nil {
		if len(alias) == 0 {
			return guestAnswer(lc.Tr(GuestHintReplyTr))
		}
		packages, err := handler.packageService.FindByAliases(uid, []string{alias})
		if err != nil {
			return handler.failure(err, "PackageService", "FindByAliases", lc.Tr(RefStatusFailure))
		}
		if len(packages) == 0 {
			return guestAnswer(lc.Tr(RefStatusNoPacks))
		}
		return guestAnswer(lc.Tr(RefStatusSuccessPackagesByAlias) + "\n\n" + buildPackagesList(packages))
	}

	aliases, err := findAliasesByObject(handler.aliasService, uid, wizard.NewAutoField(FieldObject, replyTo))
	if err != nil {
		return handler.failure(err, "AliasService", "findAliasesByObject", lc.Tr(RefStatusFailure))
	}
	if len(aliases) == 0 {
		return guestAnswer(lc.Tr(RefStatusNoFavs))
	}
	packages, err := handler.packageService.FindByAliases(uid, aliases)
	if err != nil {
		// the aliases are still useful without the packages
		log.Error("Failed to call PackageService.FindByAliases",
			logconst.FieldHandler, "GuestCommandsHandler",
			logconst.FieldMethod, "ref",
			logconst.FieldCalledObject, "PackageService",
			logconst.FieldCalledMethod, "FindByAliases",
			logconst.FieldError, err)
	}
	return guestAnswer(buildRefText(lc, aliases, packages))
}

// list shows only the first page of the favs: a guest query can be answered once, so there are no buttons for the rest.
func (handler *GuestCommandsHandler) list(lc *loc.Context, uid int64, grep string) tgbotapi.InlineQueryResult {
	page, err := handler.aliasService.ListWithCounts(uid, grep, "")
	if err != nil {
		return handler.failure(err, "AliasService", "ListWithCounts", lc.Tr(ListStatusFailure))
	}
	if len(page.Items) == 0 {
		return guestAnswer(lc.Tr(ListStatusNoRowsFavs))
	}
	text := buildText(lc.Tr(ListStatusSuccessFavs), page)
	if page.HasNextPage {
		text += "\n…"
	}
	return guestAnswer(text)
}

func (handler *GuestCommandsHandler) link(lc *loc.Context, uid int64, args string) tgbotapi.InlineQueryResult {
	name, refAlias, found := splitLinkArgs(args)
	if !found || len(name) == 0 || len(refAlias) == 0 {
		return guestMarkdownAnswer(lc.Tr(GuestLinkHintTr))
	}
	if problem := checkName(lc, name, LinkFieldTrPrefix+FieldName+FieldMaxLengthErrorTrSuffix, LinkStatusErrorForbiddenSymbolsInName); problem != "" {
		return guestAnswer(problem)
	}
	err := handler.linkService.Create(uid, name, refAlias)
	return guestAnswer(lc.Tr(linkStatus(err)))
}

func (handler *GuestCommandsHandler) rmLink(lc *loc.Context, uid int64, name string) tgbotapi.InlineQueryResult {
	if len(name) == 0 {
		return guestMarkdownAnswer(lc.Tr(GuestRmLinkHintTr))
	}
	err := handler.linkService.Delete(uid, name)
	return guestAnswer(lc.Tr(rmLinkStatus(err)))
}

// splitLinkArgs splits the arguments of the link command by the first separator found.
func splitLinkArgs(args string) (name, refAlias string, found bool) {
	for _, sep := range guestLinkSeparators {
		if name, refAlias, found = strings.Cut(args, sep); found {
			return strings.TrimSpace(name), strings.TrimSpace(refAlias), true
		}
	}
	return "", "", false
}

// checkReplyArgs checks that the alias is given and the command is a reply to a message.
// It returns the text explaining what is wrong, or an empty string.
func checkReplyArgs(lc *loc.Context, alias string, replyTo *tgbotapi.Message, hintAliasTr, hintReplyTr string) string {
	switch {
	case len(alias) == 0:
		return lc.Tr(hintAliasTr)
	case replyTo == nil:
		return lc.Tr(hintReplyTr)
	default:
		return ""
	}
}

// checkName checks the length of the name and, if forbiddenSymbolsTr isn't empty, that it has no reserved symbols.
// It returns the text explaining what is wrong, or an empty string.
func checkName(lc *loc.Context, name, maxLengthTr, forbiddenSymbolsTr string) string {
	if len([]rune(name)) > MaxAliasLen {
		return fmt.Sprintf(lc.Tr(maxLengthTr), MaxAliasLen)
	}
	if forbiddenSymbolsTr != "" {
		if err := verifyNoReservedSymbols(name, lc, forbiddenSymbolsTr); err != nil {
			return err.Error()
		}
	}
	return ""
}

func (*GuestCommandsHandler) failure(err error, service, method, text string) tgbotapi.InlineQueryResult {
	log.Error("Failed to execute a command in the guest mode",
		logconst.FieldHandler, "GuestCommandsHandler",
		logconst.FieldCalledObject, service,
		logconst.FieldCalledMethod, method,
		logconst.FieldError, err)
	return guestAnswer(text)
}

// guestAnswer wraps the text to the only result of the guest query. The first line is used as the title.
func guestAnswer(text string) tgbotapi.InlineQueryResult {
	title, _, _ := strings.Cut(text, "\n")
	return tgbotapi.NewInlineQueryResultArticle(guestAnswerResultID, title, text)
}

// guestMarkdownAnswer is like guestAnswer, but the text is sent as Markdown. Use it only for the static texts:
// the values entered by users must be escaped first.
func guestMarkdownAnswer(text string) tgbotapi.InlineQueryResult {
	article := guestAnswer(text).(tgbotapi.InlineQueryResultArticle)
	article.InputMessageContent = tgbotapi.InputTextMessageContent{Text: text, ParseMode: tgbotapi.ModeMarkdown}
	return article
}

// parseGuestCommand splits the text into the command by its keyword (the longest one wins) and the rest of the text,
// which are the arguments. It also returns the language of the keyword. It returns zero if the text doesn't start
// with a keyword.
func parseGuestCommand(text string) (cmd guestCommand, lang string, args string) {
	words := strings.Fields(strings.ToLower(text))
	for n := min(guestMaxKeywordWords, len(words)); n > 0; n-- {
		keyword := strings.Join(words[:n], " ")
		if cmd, ok := guestKeywords[keyword]; ok {
			return cmd, guestKeywordLang(keyword), dropWords(text, n)
		}
	}
	return 0, "", ""
}

// guestKeywordLang detects the language of the keyword by its alphabet.
func guestKeywordLang(keyword string) string {
	if strings.IndexFunc(keyword, func(r rune) bool { return unicode.Is(unicode.Cyrillic, r) }) >= 0 {
		return "ru"
	}
	return "en"
}

// dropWords removes n first words from the text, keeping the spaces inside the rest of it.
func dropWords(text string, n int) string {
	text = strings.TrimSpace(text)
	for ; n > 0; n-- {
		text = strings.TrimLeftFunc(text, func(r rune) bool { return !unicode.IsSpace(r) })
		text = strings.TrimLeftFunc(text, unicode.IsSpace)
	}
	return text
}
