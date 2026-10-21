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
)

const (
	RefStatusTrPrefix                 = "commands.ref.status."
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

	// only short-handed forms of commands, running in one command without the use of wizards, are supported in group chats
	if common.IsGroup(&msg.Chat) && !w.AllRequiredFieldsFilled() {
		return
	}

	w.ProcessNextField(reqenv, msg)
}

var errInvalidObjectType = errors.New("invalid type of the object field value")

func (handler *RefHandler) refAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	var (
		aliases []string
		err     error
	)
	objectField := fields.FindField(FieldObject)
	switch objectField.Type {
	case wizard.Text:
		if txt, ok := wizard.GetData[wizard.Txt](objectField); ok {
			aliases, err = handler.aliasService.ListByText(msg.From.ID, &txt)
		} else {
			err = errInvalidObjectType
		}
	case wizard.Location:
		if loc, ok := wizard.GetData[wizard.LocData](objectField); ok {
			aliases, err = handler.aliasService.ListByLocation(msg.From.ID, &loc)
		} else {
			err = errInvalidObjectType
		}
	default:
		if file, ok := wizard.GetData[wizard.File](objectField); ok {
			aliases, err = handler.aliasService.ListByFile(msg.From.ID, &file)
		} else {
			err = errInvalidObjectType
		}
	}

	replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
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

	title := reqenv.Lang.Tr(RefStatusSuccess)
	text := title + "\n\n" + LinePrefix + strings.Join(aliases, "\n"+LinePrefix)

	packages, err := handler.packageService.FindByAliases(msg.From.ID, aliases)
	if err == nil && len(packages) > 0 {
		title := reqenv.Lang.Tr(RefStatusSuccessWithPackages)
		text += "\n\n" + title + "\n\n" + buildPackagesList(packages)
	} else if err != nil {
		log.Error("Failed to call PackageService.FindByAliases",
			logconst.FieldHandler, "RefHandler",
			logconst.FieldMethod, "refAction",
			logconst.FieldCalledObject, "PackageService",
			logconst.FieldCalledMethod, "FindByAliases",
			logconst.FieldError, err)
	}

	common.ReplyPossiblySelfDestroying(handler.appenv, msg, text, base.NoOpCustomizer)
}

func buildPackagesList(packages []string) string {
	if len(packages) == 0 {
		return ""
	}
	return LinePrefix + strings.Join(packages, "\n"+LinePrefix)
}
