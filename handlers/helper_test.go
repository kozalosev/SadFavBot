package handlers

import (
	"sync"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

func buildMessage(uid int64) *tgbotapi.Message {
	return &tgbotapi.Message{
		From: &tgbotapi.User{ID: uid},
		Chat: tgbotapi.Chat{Type: "private"},
	}
}

// buildRequestEnvWithStrings returns an environment with the given English strings, since the real ones live in package main.
func buildRequestEnvWithStrings(strs map[string]string) *base.RequestEnv {
	pool := loc.NewPool("en")
	pool.Resources["en"] = strs
	reqenv := test.BuildRequestEnv()
	reqenv.Lang = pool.GetContext("en")
	return reqenv
}

var registerWizardsOnce sync.Once

// registerWizards fills the global registry of the wizard descriptors once for the whole package, since
// PopulateWizardDescriptors works only once. The questions are sent by the handler passed to NewWizard, so the tests
// see them on their own bots; only the actions of the forms run on the instances registered here.
func registerWizards() {
	registerWizardsOnce.Do(func() {
		appenv := test.BuildApplicationEnv(db)
		stateStorage := &wizard.FakeStorage{}
		start := NewStartHandler(appenv, stateStorage, StartEmbeddedHandlers{
			Language:       NewLanguageHandler(appenv, stateStorage),
			InstallPackage: NewInstallPackageHandler(appenv, stateStorage),
		})
		wizard.PopulateWizardDescriptors([]base.MessageHandler{
			start,
			start.embeddedHandlers.Language,
			start.embeddedHandlers.InstallPackage,
			NewSaveHandler(appenv, stateStorage),
			NewLinkHandler(appenv, stateStorage),
			NewDeleteHandler(appenv, stateStorage),
			NewListHandler(appenv, stateStorage),
			NewRefHandler(appenv, stateStorage),
			NewRemoveLinkHandler(appenv, stateStorage),
		})
	})
}

// buildGroupCommand builds the command sent to a supergroup.
func buildGroupCommand(uid int64, command string) *tgbotapi.Message {
	return &tgbotapi.Message{
		MessageID: 10,
		From:      &tgbotapi.User{ID: uid},
		Chat:      tgbotapi.Chat{ID: -100, Type: "supergroup"},
		Text:      command,
		Entities:  []tgbotapi.MessageEntity{{Type: "bot_command", Length: len(command)}},
	}
}
