package handlers

import (
	"errors"
	"fmt"
	log "log/slog"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"github.com/thoas/go-funk"
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
	createOrDeleteDesc.InlineKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message, form *wizard.Form) []string {
		packageName := form.Fields.FindField(FieldName).Data
		if nameField, ok := packageName.(wizard.Txt); ok {
			pkgInfo := &repo.PackageInfo{
				UID:  msg.From.ID,
				Name: nameField.Value,
			}
			if exists, err := handler.packageService.Exists(pkgInfo); err == nil {
				if exists {
					return []string{Recreate, Delete}
				}
				log.Warn("Unexpected case when a package doesn't exist but createOrDelete was requested",
					logconst.FieldHandler, "PackageHandler",
					logconst.FieldMethod, "GetWizardDescriptor")
				return []string{Create}
			} else {
				log.Error("Failed to call PackageService",
					logconst.FieldHandler, "PackageHandler",
					logconst.FieldMethod, "GetWizardDescriptor",
					logconst.FieldCalledObject, "PackageService",
					logconst.FieldError, err)
			}
		}
		return []string{Create, Recreate, Delete}
	}

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

func (*PackageHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateChats
}

func (handler *PackageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	if common.IsGroup(&msg.Chat) {
		return
	}

	name := msg.CommandArguments()

	w := wizard.NewWizard(handler, 3)
	if len(name) > 0 {
		pkgInfo := &repo.PackageInfo{
			UID:  msg.From.ID,
			Name: name,
		}
		exists, err := handler.packageService.Exists(pkgInfo)
		if err == nil && !exists {
			w.AddPrefilledField(FieldCreateOrDelete, Create)
		} else {
			w.AddEmptyField(FieldCreateOrDelete, wizard.Text)
		}
		if err != nil {
			log.Error("Failed to call PackageService.Exists",
				logconst.FieldHandler, "PackageHandler",
				logconst.FieldMethod, "Handle",
				logconst.FieldCalledObject, "PackageService",
				logconst.FieldCalledMethod, "Exists",
				logconst.FieldError, err)
		}

		w.AddPrefilledField(FieldName, name)
	} else {
		w.AddEmptyField(FieldCreateOrDelete, wizard.Text)
		w.AddEmptyField(FieldName, wizard.Text)
	}
	w.AddEmptyField(FieldAliases, wizard.Text)

	w.ProcessNextField(reqenv, msg)
}

func (handler *PackageHandler) packageAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	intent := common.TextValue(fields, FieldCreateOrDelete)
	name := strings.ReplaceAll(common.TextValue(fields, FieldName), " ", "-")

	var (
		packID string
		err    error
	)
	if intent == Delete {
		err = handler.packageService.Delete(uid, name)
	} else {
		aliasesStr := common.TextValue(fields, FieldAliases)
		aliases := strings.Split(aliasesStr, "\n")
		aliases = funk.Map(aliases, func(a string) string {
			return strings.TrimPrefix(a, LinePrefix)
		}).([]string)

		if intent == Recreate {
			packID, err = handler.packageService.Recreate(uid, name, aliases)
		} else {
			packID, err = handler.packageService.Create(uid, name, aliases)
		}
	}

	reply := base.NewReplier(handler.appenv, reqenv, msg)
	if storage.DuplicateConstraintViolation(err) {
		reply(PackageStatusDuplicate)
	} else if errors.Is(err, repo.NoRowsWereAffected) {
		reply(PackageStatusNoRows)
	} else if err != nil {
		log.Error("Failed to call PackageService",
			logconst.FieldHandler, "PackageHandler",
			logconst.FieldMethod, "packageAction",
			logconst.FieldCalledObject, "PackageService",
			logconst.FieldError, err)
		reply(PackageStatusFailure)
	} else if intent == Delete {
		reply(PackageStatusDeletionSuccess)
	} else {
		template := reqenv.Lang.Tr(PackageStatusCreationSuccess)
		packName := repo.FormatPackageName(uid, name)
		handler.appenv.Bot.ReplyWithMarkdown(msg, fmt.Sprintf(template, packName, packName, handler.appenv.Bot.GetName(), packID))
	}
}
