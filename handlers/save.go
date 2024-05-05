package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	SaveFieldsTrPrefix  = "commands.save.fields."
	SaveStatusTrPrefix  = "commands.save.status."
	SaveStatusSuccess   = SaveStatusTrPrefix + StatusSuccess
	SaveStatusFailure   = SaveStatusTrPrefix + StatusFailure
	SaveStatusDuplicate = SaveStatusTrPrefix + StatusDuplicate

	SaveStatusErrorForbiddenSymbolsInAlias = SaveFieldsTrPrefix + FieldAlias + FieldValidationErrorTrInfix + "forbidden.symbols"
	SaveObjectValidationErrorCustomEmoji   = SaveFieldsTrPrefix + FieldObject + FieldValidationErrorTrInfix + "custom.emoji"

	MaxAliasLen               = 128
	MaxTextLen                = 4096
	ReservedSymbols           = ReservedSymbolsForMessage + "\n"
	ReservedSymbolsForMessage = "•@|{}[]:"
)

type SaveHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	favService *repo.FavService
}

func NewSaveHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *SaveHandler {
	h := &SaveHandler{
		appenv:       appenv,
		stateStorage: stateStorage,
		favService:   repo.NewFavsService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *SaveHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *SaveHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.saveFormAction)

	aliasDesc := desc.AddField(FieldAlias, SaveFieldsTrPrefix+FieldAlias)
	aliasDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len([]rune(msg.Text)) > MaxAliasLen {
			template := lc.Tr(SaveFieldsTrPrefix + FieldAlias + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, MaxAliasLen))
		}
		return verifyNoReservedSymbols(msg.Text, lc, SaveStatusErrorForbiddenSymbolsInAlias)
	}

	objDesc := desc.AddField(FieldObject, SaveFieldsTrPrefix+FieldObject)
	objDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len([]rune(msg.Text)) > MaxTextLen {
			template := lc.Tr(SaveFieldsTrPrefix + FieldObject + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, MaxTextLen))
		}
		for _, entity := range msg.Entities {
			if len(entity.CustomEmojiID) > 0 {
				return errors.New(fmt.Sprintf(lc.Tr(SaveObjectValidationErrorCustomEmoji)))
			}
		}
		return nil
	}

	return desc
}

func (*SaveHandler) GetCommands() []string {
	return saveCommands
}

func (*SaveHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateAndGroupChats
}

func (handler *SaveHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	wizardForm := wizard.NewWizard(handler, 2)
	title := msg.CommandArguments()
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

	if msg.ReplyToMessage != nil {
		wizardForm.AddPrefilledAutoField(FieldObject, msg.ReplyToMessage)
	} else {
		wizardForm.AddEmptyField(FieldObject, wizard.Auto)
	}

	// only short-handed forms of commands, running in one command without the use of wizards, are supported in group chats
	if common.IsGroup(&msg.Chat) && !wizardForm.AllRequiredFieldsFilled() {
		return
	}

	wizardForm.ProcessNextField(reqenv, msg)
}

func (handler *SaveHandler) saveFormAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	alias, fav := common.ExtractFavInfo(fields)

	replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
	if len(alias) == 0 {
		replyWith(SaveStatusFailure)
		return
	}

	res, err := handler.favService.Save(uid, alias, fav)

	if err != nil {
		if storage.DuplicateConstraintViolation(err) {
			replyWith(SaveStatusDuplicate)
		} else {
			log.WithField(logconst.FieldHandler, "SaveHandler").
				WithField(logconst.FieldMethod, "saveFormAction").
				WithField(logconst.FieldCalledObject, "FavService").
				WithField(logconst.FieldCalledMethod, "Save").
				Error(err)
			replyWith(SaveStatusFailure)
		}
	} else {
		if res.RowsAffected() > 0 {
			answer := fmt.Sprintf(reqenv.Lang.Tr(SaveStatusSuccess), handler.appenv.Bot.GetName(), common.MarkdownEscaper.Replace(alias))
			common.ReplyPossiblySelfDestroying(handler.appenv, msg, answer, base.MarkdownCustomizer)
		} else {
			log.WithField(logconst.FieldHandler, "SaveHandler").
				WithField(logconst.FieldMethod, "saveFormAction").
				Warning("No rows were affected!")
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
