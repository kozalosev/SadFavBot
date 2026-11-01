package help

import (
	_ "embed"
	"strings"
)

//go:embed guest.md
var guestHelpEn string

//go:embed guest.ru.md
var guestHelpRu string

// GuestHelp returns the help messages for the guest mode in English and Russian.
func GuestHelp() (en, ru string) {
	return strings.TrimSpace(guestHelpEn), strings.TrimSpace(guestHelpRu)
}
