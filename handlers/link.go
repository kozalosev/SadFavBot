package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
)

const (
	LinkFieldTrPrefix                     = "commands.link.fields."
	LinkStatusTrPrefix                    = "commands.link.status."
	LinkStatusSuccess                     = LinkStatusTrPrefix + StatusSuccess
	LinkStatusFailure                     = LinkStatusTrPrefix + StatusFailure
	LinkStatusDuplicate                   = LinkStatusTrPrefix + StatusDuplicate
	LinkStatusDuplicateFav                = LinkStatusTrPrefix + StatusDuplicate + ".fav"
	LinkStatusNoAlias                     = LinkStatusTrPrefix + "no.alias"
	LinkStatusErrorForbiddenSymbolsInName = LinkFieldTrPrefix + FieldName + FieldValidationErrorTrInfix + "forbidden.symbols"
)

type LinkHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	linkService  *repo.LinkService
	aliasService *repo.AliasService
}

func NewLinkHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *LinkHandler {
	h := &LinkHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		linkService:  repo.NewLinkService(appenv),
		aliasService: repo.NewAliasService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *LinkHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *LinkHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.linkAction)

	nameField := desc.AddField(FieldName, LinkFieldTrPrefix+FieldName)
	nameField.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxAliasLen {
			template := lc.Tr(LinkFieldTrPrefix + FieldName + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return verifyNoReservedSymbols(msg.Text, lc, LinkStatusErrorForbiddenSymbolsInName)
	}

	aliasField := desc.AddField(FieldAlias, LinkFieldTrPrefix+FieldAlias)
	aliasField.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		aliases, err := handler.aliasService.ListForFavsOnly(msg.From.ID)
		if err != nil {
			log.WithField(logconst.FieldHandler, "LinkHandler").
				WithField(logconst.FieldFunc, "ReplyKeyboardBuilder").
				WithField(logconst.FieldCalledObject, "AliasService").
				WithField(logconst.FieldCalledMethod, "ListForFavsOnly").
				Error(err)
		}
		return aliases
	}

	return desc
}

func (*LinkHandler) GetCommands() []string {
	return linkCommands
}

func (handler *LinkHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 2)
	if name := base.GetCommandArgument(msg); len(name) > 0 {
		argParts := funk.Map(strings.Split(name, "->"), func(s string) string {
			return strings.TrimSpace(s)
		}).([]string)
		if len(argParts) == 2 {
			if len(argParts[0]) <= MaxAliasLen && verifyNoReservedSymbols(argParts[0], reqenv.Lang, LinkStatusErrorForbiddenSymbolsInName) == nil {
				w.AddPrefilledField(FieldName, argParts[0])
			} else {
				w.AddEmptyField(FieldName, wizard.Text)
			}
			w.AddPrefilledField(FieldAlias, argParts[1])
		} else {
			if len(name) <= MaxAliasLen && verifyNoReservedSymbols(name, reqenv.Lang, LinkStatusErrorForbiddenSymbolsInName) == nil {
				w.AddPrefilledField(FieldName, name)
			} else {
				w.AddEmptyField(FieldName, wizard.Text)
			}
			w.AddEmptyField(FieldAlias, wizard.Text)
		}
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
		w.AddEmptyField(FieldAlias, wizard.Text)
	}
	w.ProcessNextField(reqenv, msg)
}

func (handler *LinkHandler) linkAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	name := fields.FindField(FieldName).Data.(string)
	refAlias := fields.FindField(FieldAlias).Data.(string)

	err := handler.linkService.Create(uid, name, refAlias)

	reply := base.NewReplier(handler.appenv, reqenv, msg)
	if isAttemptToInsertLinkForExistingFav(err) {
		reply(LinkStatusDuplicateFav)
	} else if storage.DuplicateConstraintViolation(err) {
		reply(LinkStatusDuplicate)
	} else if isAttemptToLinkNonExistingAlias(err) {
		reply(LinkStatusNoAlias)
	} else if err != nil {
		log.WithField(logconst.FieldHandler, "LinkHandler").
			WithField(logconst.FieldMethod, "linkAction").
			WithField(logconst.FieldCalledObject, "LinkService").
			WithField(logconst.FieldCalledMethod, "Create").
			Error(err)
		reply(LinkStatusFailure)
	} else {
		reply(LinkStatusSuccess)
	}
}

func isAttemptToInsertLinkForExistingFav(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Message == "Insertion of the link with the same name as an already existing fav is forbidden"
}

func isAttemptToLinkNonExistingAlias(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23502" && pgErr.ColumnName == "linked_alias_id"
}
