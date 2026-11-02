package handlers

import (
	"errors"
	log "log/slog"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

const (
	RefStatusTrPrefix                 = "commands.ref.status."
	RefGroupHintTr                    = "commands.ref.group.hint"
	RefStatusSuccess                  = RefStatusTrPrefix + StatusSuccess
	RefStatusSuccessWithPackages      = RefStatusTrPrefix + StatusSuccess + ".with.packages"
	RefStatusSuccessPackagesByAlias   = RefStatusTrPrefix + StatusSuccess + ".packages.by.alias"
	RefStatusFailure                  = RefStatusTrPrefix + StatusFailure
	RefStatusNoFavs                   = RefStatusTrPrefix + "no.favs"
	RefStatusNoPacks                  = RefStatusTrPrefix + "no.packages"
	RefFieldAliasesOrPackagesPromptTr = "commands.ref.fields.object"
)

type RefHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appenv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	aliasService   *repo.AliasService
	packageService *repo.PackageService
}

func NewRefHandler(appenv *base.ApplicationEnv, stateStorage wizard.StateStorage) *RefHandler {
	h := &RefHandler{
		appenv:         appenv,
		stateStorage:   stateStorage,
		aliasService:   repo.NewAliasService(appenv),
		packageService: repo.NewPackageService(appenv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (handler *RefHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(handler.appenv, handler.stateStorage)
}

func (handler *RefHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(handler.refAction)
	desc.AddField(FieldObject, RefFieldAliasesOrPackagesPromptTr)
	return desc
}

func (*RefHandler) GetCommands() []string {
	return refCommands
}

// IsEphemeral makes the command ephemeral in group chats: nobody except the sender sees the reply.
func (*RefHandler) IsEphemeral() bool {
	return true
}

func (handler *RefHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	if len(msg.CommandArguments()) > 0 {
		alias := msg.CommandArguments()
		packages, err := handler.packageService.FindByAliases(msg.From.ID, []string{alias})

		replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
		if err != nil {
			log.Error("Failed to call PackageService.FindByAliases",
				logconst.FieldHandler, "RefHandler",
				logconst.FieldMethod, "Handle",
				logconst.FieldCalledObject, "PackageService",
				logconst.FieldCalledMethod, "FindByAliases",
				logconst.FieldError, err)
			replyWith(RefStatusFailure)
		} else if len(packages) == 0 {
			replyWith(RefStatusNoPacks)
		} else {
			text := reqenv.Lang.Tr(RefStatusSuccessPackagesByAlias) + "\n\n" + buildPackagesList(packages)
			common.ReplyPossiblySelfDestroying(handler.appenv, msg, text, base.NoOpCustomizer)
		}
		return
	}

	w := wizard.NewWizard(handler, 1)
	if msg.ReplyToMessage != nil {
		w.AddPrefilledAutoField(FieldObject, msg.ReplyToMessage)
	} else {
		w.AddEmptyField(FieldObject, wizard.Auto)
	}

	// the command is ephemeral in groups, and the user can't answer an ephemeral question, so only the short form works
	if base.IsGroupChat(&msg.Chat) && !w.AllRequiredFieldsFilled() {
		common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)(RefGroupHintTr)
		return
	}

	w.ProcessNextField(reqenv, msg)
}

var errInvalidObjectType = errors.New("invalid type of the object field value")

func (handler *RefHandler) refAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	msg = common.ReplyToOriginal(msg)
	replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
	objectField := fields.FindField(FieldObject)
	if objectField == nil {
		log.Error("Object field not found",
			logconst.FieldHandler, "RefHandler",
			logconst.FieldMethod, "refAction")
		replyWith(RefStatusFailure)
		return
	}

	aliases, err := findAliasesByObject(handler.aliasService, msg.From.ID, objectField)
	if errors.Is(err, errInvalidObjectType) {
		log.Error("Invalid type of the object field value",
			logconst.FieldHandler, "RefHandler",
			logconst.FieldMethod, "refAction",
			logconst.FieldValue, objectField.Data)
		replyWith(RefStatusFailure)
		return
	} else if err != nil {
		log.Error("Failed to call AliasService",
			logconst.FieldHandler, "RefHandler",
			logconst.FieldMethod, "refAction",
			logconst.FieldCalledObject, "AliasService",
			logconst.FieldError, err)
		replyWith(RefStatusFailure)
		return
	} else if len(aliases) == 0 {
		replyWith(RefStatusNoFavs)
		return
	}

	packages, err := handler.packageService.FindByAliases(msg.From.ID, aliases)
	if err != nil {
		log.Error("Failed to call PackageService.FindByAliases",
			logconst.FieldHandler, "RefHandler",
			logconst.FieldMethod, "refAction",
			logconst.FieldCalledObject, "PackageService",
			logconst.FieldCalledMethod, "FindByAliases",
			logconst.FieldError, err)
	}

	text := buildRefText(reqenv.Lang, aliases, packages)
	common.ReplyPossiblySelfDestroying(handler.appenv, msg, text, base.NoOpCustomizer)
}

// findAliasesByObject returns all aliases of the user the object is saved under.
func findAliasesByObject(service *repo.AliasService, uid int64, objectField *wizard.Field) ([]string, error) {
	switch objectField.Type {
	case wizard.Text:
		if txt, ok := objectField.Data.(wizard.Txt); ok {
			return service.ListByText(uid, &txt)
		}
	case wizard.Location:
		if loc, ok := objectField.Data.(wizard.LocData); ok {
			return service.ListByLocation(uid, &loc)
		}
	default:
		if file, ok := objectField.Data.(wizard.File); ok {
			return service.ListByFile(uid, &file)
		}
	}
	return nil, errInvalidObjectType
}

// buildRefText builds the answer with the aliases of an object and the packages containing them, if any.
func buildRefText(lc *loc.Context, aliases, packages []string) string {
	text := lc.Tr(RefStatusSuccess) + "\n\n" + LinePrefix + strings.Join(aliases, "\n"+LinePrefix)
	if len(packages) > 0 {
		text += "\n\n" + lc.Tr(RefStatusSuccessWithPackages) + "\n\n" + buildPackagesList(packages)
	}
	return text
}

func buildPackagesList(packages []string) string {
	if len(packages) == 0 {
		return ""
	}
	return LinePrefix + strings.Join(packages, "\n"+LinePrefix)
}
