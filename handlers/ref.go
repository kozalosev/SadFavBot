package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
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

func (*RefHandler) GetScopes() []base.CommandScope {
	return common.CommandScopePrivateAndGroupChats
}

func (handler *RefHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	if len(msg.CommandArguments()) > 0 {
		alias := msg.CommandArguments()
		packages, err := handler.packageService.FindByAliases(msg.From.ID, []string{alias})

		replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
		if err != nil {
			log.WithField(logconst.FieldHandler, "RefHandler").
				WithField(logconst.FieldMethod, "Handle").
				WithField(logconst.FieldCalledObject, "PackageService").
				WithField(logconst.FieldCalledMethod, "FindByAliases").
				Error(err)
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
	if common.IsGroup(msg.Chat) && !w.AllRequiredFieldsFilled() {
		return
	}

	w.ProcessNextField(reqenv, msg)
}

func (handler *RefHandler) refAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	var (
		aliases []string
		err     error
	)
	objField := fields.FindField(FieldObject)
	switch objField.Type {
	case wizard.Text:
		txt := objField.Data.(wizard.Txt)
		aliases, err = handler.aliasService.ListByText(msg.From.ID, &txt)
	case wizard.Location:
		loc := objField.Data.(wizard.LocData)
		aliases, err = handler.aliasService.ListByLocation(msg.From.ID, &loc)
	default:
		file := objField.Data.(wizard.File)
		aliases, err = handler.aliasService.ListByFile(msg.From.ID, &file)
	}

	replyWith := common.PossiblySelfDestroyingReplier(handler.appenv, reqenv, msg)
	if err != nil {
		log.WithField(logconst.FieldHandler, "RefHandler").
			WithField(logconst.FieldMethod, "refAction").
			WithField(logconst.FieldCalledObject, "AliasService").
			Error(err)
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
		log.WithField(logconst.FieldHandler, "RefHandler").
			WithField(logconst.FieldMethod, "refAction").
			WithField(logconst.FieldCalledObject, "PackageService").
			WithField(logconst.FieldCalledMethod, "FindByAliases").
			Error(err)
	}

	common.ReplyPossiblySelfDestroying(handler.appenv, msg, text, base.NoOpCustomizer)
}

func buildPackagesList(packages []string) string {
	if len(packages) == 0 {
		return ""
	}
	return LinePrefix + strings.Join(packages, "\n"+LinePrefix)
}
