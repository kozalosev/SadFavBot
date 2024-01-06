package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	PackageFieldsTrPrefix        = "commands.package.fields."
	PackageStatusTrPrefix        = "commands.package.status."
	PackageStatusCreationSuccess = PackageStatusTrPrefix + StatusSuccess + ".creation"
	PackageStatusDeletionSuccess = PackageStatusTrPrefix + StatusSuccess + ".deletion"
	PackageStatusFailure         = PackageStatusTrPrefix + StatusFailure
	PackageStatusDuplicate       = PackageStatusTrPrefix + StatusDuplicate
	PackageStatusNoRows          = PackageStatusTrPrefix + StatusNoRows

	PackageStatusErrorForbiddenSymbolsInName = PackageFieldsTrPrefix + FieldName + FieldValidationErrorTrInfix + "forbidden.symbols"

	FieldCreateOrDelete = "createOrDelete"
	FieldName           = "name"
	FieldAliases        = FieldAlias + "es"

	Create   = "Create"
	Recreate = "Recreate"
	Delete   = "Delete"

	MaxPackageNameLen = 256
)

type PackageHandler struct {
	base.CommandHandlerTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	packageService *repo.PackageService
}

func NewPackageHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *PackageHandler {
	h := &PackageHandler{
		appenv:         appenv,
		stateStorage:   stateStorage,
		packageService: repo.NewPackageService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *PackageHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *PackageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.packageAction)

	createOrDeleteDesc := desc.AddField(FieldCreateOrDelete, PackageFieldsTrPrefix+FieldCreateOrDelete)
	createOrDeleteDesc.InlineKeyboardAnswers = []string{Create, Recreate, Delete}

	nameDesc := desc.AddField(FieldName, PackageFieldsTrPrefix+FieldName)
	nameDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len([]rune(msg.Text)) > MaxPackageNameLen {
			template := lc.Tr(PackageFieldsTrPrefix + FieldName + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, MaxPackageNameLen))
		}
		return verifyNoReservedSymbols(msg.Text, lc, PackageStatusErrorForbiddenSymbolsInName)
	}

	aliasesDesc := desc.AddField(FieldAliases, PackageFieldsTrPrefix+FieldAliases)
	aliasesDesc.SkipIf = &wizard.SkipOnFieldValue{
		Name:  FieldCreateOrDelete,
		Value: Delete,
	}
	return desc
}

func (*PackageHandler) GetCommands() []string {
	return packageCommands
}

func (handler *PackageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	name := base.GetCommandArgument(msg)

	w := wizard.NewWizard(handler, 3)
	w.AddEmptyField(FieldCreateOrDelete, wizard.Text)
	if len(name) > 0 {
		w.AddPrefilledField(FieldName, name)
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
	}
	w.AddEmptyField(FieldAliases, wizard.Text)

	w.ProcessNextField(reqenv, msg)
}

func (handler *PackageHandler) packageAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	intent := fields.FindField(FieldCreateOrDelete).Data.(wizard.Txt).Value
	name := strings.ReplaceAll(fields.FindField(FieldName).Data.(wizard.Txt).Value, " ", "-")

	var err error
	if intent == Delete {
		err = handler.packageService.Delete(uid, name)
	} else {
		aliasesStr := fields.FindField(FieldAliases).Data.(wizard.Txt).Value
		aliases := strings.Split(aliasesStr, "\n")
		aliases = funk.Map(aliases, func(a string) string {
			return strings.TrimPrefix(a, LinePrefix)
		}).([]string)

		if intent == Recreate {
			err = handler.packageService.Recreate(uid, name, aliases)
		} else {
			err = handler.packageService.Create(uid, name, aliases)
		}
	}

	reply := base.NewReplier(handler.appenv, reqenv, msg)
	if storage.DuplicateConstraintViolation(err) {
		reply(PackageStatusDuplicate)
	} else if errors.Is(err, repo.NoRowsWereAffected) {
		reply(PackageStatusNoRows)
	} else if err != nil {
		log.WithField(logconst.FieldHandler, "PackageHandler").
			WithField(logconst.FieldMethod, "packageAction").
			WithField(logconst.FieldCalledObject, "PackageService").
			Error(err)
		reply(PackageStatusFailure)
	} else if intent == Delete {
		reply(PackageStatusDeletionSuccess)
	} else {
		template := reqenv.Lang.Tr(PackageStatusCreationSuccess)
		packName := repo.FormatPackageName(uid, name)
		urlPath := base64.URLEncoding.EncodeToString([]byte(packName))
		handler.appenv.Bot.ReplyWithMarkdown(msg, fmt.Sprintf(template, packName, packName, handler.appenv.Bot.GetName(), urlPath))
	}
}
