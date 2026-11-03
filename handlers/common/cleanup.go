package common

import (
	"fmt"
	log "log/slog"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/selfdestruct"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

// The cleanup of group chats: where the bot may delete messages of other users, its replies and the questions of the
// wizards disappear together with the commands and the answers of the users, after a warning; elsewhere, nothing is
// deleted, so that no lone messages are left.

// SelfDestructionWarningTr is shown in place of the messages of the bot before they're deleted.
const SelfDestructionWarningTr = "self.destruction.warning"

var (
	// selfDeletionWarning is how long the warning is shown before the deletion; zero deletes without warning.
	selfDeletionWarning = secondsFromEnv("SELF_DELETION_WARNING_SECS", 10)
	// selfDeletionCancelDelay replaces the delay after /cancel: there is nothing to read there.
	selfDeletionCancelDelay = secondsFromEnv("SELF_DELETION_CANCEL_DELAY_SECS", 5)
)

func secondsFromEnv(name string, defaultSecs int64) time.Duration {
	secs := defaultSecs
	if s, ok := os.LookupEnv(name); ok {
		if parsed, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Error("Failed to parse an environment variable",
				logconst.FieldFunc, "secondsFromEnv",
				logconst.FieldConst, name,
				logconst.FieldError, err)
		} else {
			secs = parsed
		}
	}
	return time.Duration(secs) * time.Second
}

// QuickDeletion makes a self-destroying reply disappear after the reduced delay of the cancellation.
func QuickDeletion(o *replyOptions) { o.delay = selfDeletionCancelDelay }

// GroupCleanupPolicy deletes the whole dialogue of a wizard where the bot may delete messages of other users:
// the delay of the self-destroying messages after it's completed, the reduced delay after it's cancelled, or
// abandonedAfter the last question nobody has answered (pass the lifetime of the state of a form).
func GroupCleanupPolicy(abandonedAfter time.Duration) wizard.GroupCleanupPolicy {
	cleanupIfAllowed := wizard.CleanupIfAllowed(selfDeletionDelay, abandonedAfter)
	return func(d wizard.GroupDialogue) wizard.GroupCleanup {
		cleanup := cleanupIfAllowed(d)
		if len(cleanup.Delete) > 0 {
			// the last question is the one the users see; Telegram edits one message per request
			cleanup.Warning = deletionWarning(d.Lang, d.QuestionIDs[len(d.QuestionIDs)-1])
		}
		if d.Stage == wizard.StageCancelled {
			cleanup.After = selfDeletionCancelDelay
		}
		return cleanup
	}
}

// deletionWarning returns the warning to show in place of the messages of the bot, or nil if it's disabled.
func deletionWarning(lang *loc.Context, botMessageID base.MessageID) *selfdestruct.Warning {
	if selfDeletionWarning <= 0 {
		return nil
	}
	return &selfdestruct.Warning{
		Text:      fmt.Sprintf(lang.Tr(SelfDestructionWarningTr), int(selfDeletionWarning.Seconds())),
		ParseMode: tgbotapi.ModeHTML,
		MessageID: botMessageID,
		Grace:     selfDeletionWarning,
	}
}
