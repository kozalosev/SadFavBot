package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/loctools/go-l10n/loc"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"net/url"
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

	Create = "Create"
	Delete = "Delete"

	MaxPackageNameLen = 256
)

type PackageHandler struct {
	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	packageService *repo.PackageService
}

func NewPackageHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) PackageHandler {
	return PackageHandler{
		appenv:         appenv,
		stateStorage:   stateStorage,
		packageService: repo.NewPackageService(appenv),
	}
}

func (handler PackageHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler PackageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.packageAction)

	createOrDeleteDesc := desc.AddField(FieldCreateOrDelete, PackageFieldsTrPrefix+FieldCreateOrDelete)
	createOrDeleteDesc.InlineKeyboardAnswers = []string{Create, Delete}

	nameDesc := desc.AddField(FieldName, PackageFieldsTrPrefix+FieldName)
	nameDesc.Validator = func(msg *tgbotapi.Message, lc *loc.Context) error {
		if len(msg.Text) > MaxPackageNameLen {
			template := lc.Tr(PackageFieldsTrPrefix + FieldName + FieldMaxLengthErrorTrSuffix)
			return errors.New(fmt.Sprintf(template, maxAliasLenStr))
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

func (PackageHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "package" || msg.Command() == "pack"
}

func (handler PackageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
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

func (handler PackageHandler) packageAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	deletion := fields.FindField(FieldCreateOrDelete).Data == Delete
	name := strings.ReplaceAll(fields.FindField(FieldName).Data.(string), " ", "-")

	var err error
	if deletion {
		err = handler.packageService.Delete(uid, name)
	} else {
		aliasesStr := fields.FindField(FieldAliases).Data.(string)
		aliases := strings.Split(aliasesStr, "\n")
		aliases = funk.Map(aliases, func(a string) string {
			return strings.TrimPrefix(a, LinePrefix)
		}).([]string)

		err = handler.packageService.Create(uid, name, aliases)
	}

	reply := replierFactory(handler.appenv, reqenv, msg)
	if isDuplicateConstraintViolation(err) {
		reply(PackageStatusDuplicate)
	} else if err == repo.NoRowsWereAffected {
		reply(PackageStatusNoRows)
	} else if err != nil {
		log.Error(err)
		reply(PackageStatusFailure)
	} else if deletion {
		reply(PackageStatusDeletionSuccess)
	} else {
		template := reqenv.Lang.Tr(PackageStatusCreationSuccess)
		packName := repo.FormatPackageName(uid, name)
		urlPath := url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(packName)))
		handler.appenv.Bot.ReplyWithMarkdown(msg, fmt.Sprintf(template, packName, packName, handler.appenv.Bot.GetName(), urlPath))
	}
}
