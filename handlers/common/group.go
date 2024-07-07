package common

import "github.com/kozalosev/goSadTgBot/base"

type GroupCommand interface {
	base.CommandHandler

	isGroupCommand() bool
}

type GroupCommandTrait struct{}

func (*GroupCommandTrait) isGroupCommand() bool {
	return true
}

func (*GroupCommandTrait) GetScopes() []base.CommandScope {
	return CommandScopePrivateAndGroupChats
}
