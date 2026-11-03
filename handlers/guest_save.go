package handlers

import (
	log "log/slog"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

const (
	GuestSaveHintAliasTr   = "guest.save.hint.alias"
	GuestSaveHintReplyTr   = "guest.save.hint.reply"
	GuestSaveUnsupportedTr = "guest.save.unsupported"
	GuestSaveSuccessTr     = "guest.save.success"
)

// save saves the replied message under the alias: "@bot save alias".
func (handler *GuestCommandsHandler) save(lc *loc.Context, uid int64, replyTo *tgbotapi.Message, alias string) tgbotapi.InlineQueryResult {
	if problem := checkReplyArgs(lc, alias, replyTo, GuestSaveHintAliasTr, GuestSaveHintReplyTr); problem != "" {
		return guestMarkdownAnswer(problem)
	}
	if problem := checkName(lc, alias, SaveFieldsTrPrefix+FieldAlias+FieldMaxLengthErrorTrSuffix, SaveStatusErrorForbiddenSymbolsInAlias); problem != "" {
		return guestAnswer(problem)
	}
	fav := favFromMessage(replyTo)
	if fav == nil {
		return guestAnswer(lc.Tr(GuestSaveUnsupportedTr))
	}

	// the user may have never started the bot
	if _, err := handler.userService.Create(uid); err != nil {
		return handler.failure(err, "UserService", "Create", lc.Tr(SaveStatusFailure))
	}
	res, err := handler.favService.Save(uid, alias, fav)
	countSaveOutcome(metricModeGuest, res, err)
	if err != nil {
		if storage.DuplicateConstraintViolation(err) {
			return guestAnswer(lc.Tr(SaveStatusDuplicate))
		}
		return handler.failure(err, "FavService", "Save", lc.Tr(SaveStatusFailure))
	}
	if res.RowsAffected() == 0 {
		log.Warn("No rows were affected!",
			logconst.FieldHandler, "GuestCommandsHandler",
			logconst.FieldMethod, "save")
		return guestAnswer(lc.Tr(SaveStatusFailure))
	}
	return guestHTMLAnswer(common.FormatHTML(lc.Tr(GuestSaveSuccessTr), handler.appenv.Bot.GetName(), alias))
}

// favFromMessage extracts the object to save from the message by the same rules as in the wizard of the /save command.
// It returns nil for unsupported content: the message must be either a file, a location, or a text of valid length.
func favFromMessage(msg *tgbotapi.Message) *dto.Fav {
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Type: wizard.Text, Data: wizard.Txt{}},
		wizard.NewAutoField(FieldObject, msg),
	}
	_, fav := common.ExtractFavInfo(fields)
	if fav == nil {
		return nil
	}
	if fav.Type == wizard.Text && (len(fav.Text.Value) == 0 || len([]rune(fav.Text.Value)) > MaxTextLen) {
		return nil
	}
	return fav
}
