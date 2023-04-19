package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	SaveFieldsTrPrefix  = "commands.save.fields."
	SaveStatusTrPrefix  = "commands.save.status."
	SaveStatusSuccess   = SaveStatusTrPrefix + StatusSuccess
	SaveStatusFailure   = SaveStatusTrPrefix + StatusFailure
	SaveStatusDuplicate = SaveStatusTrPrefix + StatusDuplicate

	SaveStatusErrorForbiddenSymbolsInAlias = SaveFieldsTrPrefix + FieldAlias + FieldValidationErrorTrInfix + "forbidden.symbols"

	MaxAliasLen               = 128
	MaxTextLen                = 4096
	ReservedSymbols           = ReservedSymbolsForMessage + "\n"
	ReservedSymbolsForMessage = "•@|{}[]:"
)

var (
	maxAliasLenStr = strconv.FormatInt(MaxAliasLen, 10)
	maxTextLenStr  = strconv.FormatInt(MaxAliasLen, 10)
)

type SaveHandler struct {
	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	favService *repo.FavService
}

func NewSaveHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) SaveHandler {
	return SaveHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		favService:   repo.NewFavsService(appenv),
	}
}

func (handler SaveHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler SaveHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.saveFormAction)

	aliasDesc := desc.AddField(FieldAlias, SaveFieldsTrPrefix+FieldAlias)
	aliasDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxAliasLen {
			template := lc.Tr(SaveFieldsTrPrefix + FieldAlias + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
		}
		return verifyNoReservedSymbols(msg.Text, lc, SaveStatusErrorForbiddenSymbolsInAlias)
	}

	objDesc := desc.AddField(FieldObject, SaveFieldsTrPrefix+FieldObject)
	objDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxTextLen {
			template := lc.Tr(SaveFieldsTrPrefix + FieldObject + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxTextLenStr))
		}
		return nil
	}

	return desc
}

func (SaveHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "save"
}

func (handler SaveHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	wizardForm := wizard.NewWizard(handler, 2)
	title := base.GetCommandArgument(msg)
	if len(title) > 0 {
		if err := verifyNoReservedSymbols(title, reqenv.Lang, SaveStatusErrorForbiddenSymbolsInAlias); err != nil {
			handler.appenv.Bot.ReplyWithMarkdown(msg, err.Error())
			wizardForm.AddEmptyField(FieldAlias, wizard.Text)
		} else {
			wizardForm.AddPrefilledField(FieldAlias, title)
		}
	} else {
		wizardForm.AddEmptyField(FieldAlias, wizard.Text)
	}
	wizardForm.AddEmptyField(FieldObject, wizard.Auto)
	wizardForm.ProcessNextField(reqenv, msg)
}

func (handler SaveHandler) saveFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	alias, fav := extractFavInfo(fields)

	replyWith := replierFactory(handler.appenv, reqenv, msg)
	if len(alias) == 0 {
		replyWith(SaveStatusFailure)
		return
	}

	res, err := handler.favService.Save(uid, alias, fav)

	if err != nil {
		if isDuplicateConstraintViolation(err) {
			replyWith(SaveStatusDuplicate)
		} else {
			log.Errorln(err.Error())
			replyWith(SaveStatusFailure)
		}
	} else {
		if res.RowsAffected() > 0 {
			answer := fmt.Sprintf(reqenv.Lang.Tr(SaveStatusSuccess), handler.appenv.Bot.GetName(), markdownEscaper.Replace(alias))
			handler.appenv.Bot.ReplyWithMarkdown(msg, answer)
		} else {
			log.Warning("No rows were affected!")
			replyWith(SaveStatusFailure)
		}
	}
}

func verifyNoReservedSymbols(text string, lc *loc.Context, errTemplateName string) error {
	if strings.ContainsAny(text, ReservedSymbols) {
		template := lc.Tr(errTemplateName)
		return errors.New(fmt.Sprintf(template, ReservedSymbolsForMessage))
	} else {
		return nil
	}
}
