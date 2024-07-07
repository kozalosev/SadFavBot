package handlers

// List of all commands supported by the bot.
// See their descriptions in strings.go.
var (
	cancelCommands = []string{"cancel"}
	deleteCommands = []string{"delete", "del"}
	// helpCommands	   in handlers/help/command.go
	// privacyCommands in handlers/privacy/privacy.go
	installCommands    = []string{"install"}
	languageCommands   = []string{"language", "lang"}
	linkCommands       = []string{"link", "ln"}
	rmLinkCommands     = []string{"rmlink", "delink", "dellink", "remove_link", "delete_link"}
	listCommands       = []string{"list"}
	modeCommands       = []string{"mode", "mod"}
	packageCommands    = []string{"package", "pack"}
	refCommands        = []string{"ref", "references"}
	saveCommands       = []string{"save"}
	visibilityCommands = []string{"visibility", "vis"}
)
