package handlers

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

const (
	LinkFieldTrPrefix   = "commands.link.fields."
	LinkStatusTrPrefix  = "commands.link.status."
	LinkStatusSuccess   = LinkStatusTrPrefix + StatusSuccess
	LinkStatusFailure   = LinkStatusTrPrefix + StatusFailure
	LinkStatusDuplicate = LinkStatusTrPrefix + StatusDuplicate
)

type LinkHandler struct {
	StateStorage wizard.StateStorage
}

func (LinkHandler) GetWizardName() string                              { return "LinkWizard" }
func (handler LinkHandler) GetWizardStateStorage() wizard.StateStorage { return handler.StateStorage }

func (handler LinkHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(linkAction)

	desc.AddField(FieldName, LinkFieldTrPrefix+FieldName)
	aliasField := desc.AddField(FieldAlias, LinkFieldTrPrefix+FieldAlias)
	aliasField.ReplyKeyboardBuilder = func(reqenv *base.RequestEnv, msg *tgbotapi.Message) []string {
		var (
			aliases []string
			alias   string
		)
		if res, err := reqenv.Database.QueryContext(reqenv.Ctx, "SELECT DISTINCT a.name FROM items i JOIN aliases a on a.id = i.alias WHERE i.uid = $1", msg.From.ID); err == nil {
			for res.Next() {
				if err := res.Scan(&alias); err == nil {
					aliases = append(aliases, alias)
				} else {
					log.Error(err)
				}
			}
		} else {
			log.Error(err)
		}
		return aliases
	}

	return desc
}

func (handler LinkHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "link" || msg.Command() == "ln"
}

func (handler LinkHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	w := wizard.NewWizard(handler, 2)
	if name := base.GetCommandArgument(msg); len(name) > 0 {
		w.AddPrefilledField(FieldName, name)
	} else {
		w.AddEmptyField(FieldName, wizard.Text)
	}
	w.AddEmptyField(FieldAlias, wizard.Text)
	w.ProcessNextField(reqenv, msg)
}

func linkAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	uid := msg.From.ID
	name := fields.FindField(FieldName).Data.(string)
	refAlias := fields.FindField(FieldAlias).Data.(string)

	var (
		tx  *sql.Tx
		err error
	)
	if tx, err = reqenv.Database.BeginTx(reqenv.Ctx, nil); err == nil {
		var aliasID int
		if aliasID, err = saveAliasToSeparateTable(reqenv.Ctx, tx, name); err == nil {
			if _, err = tx.ExecContext(reqenv.Ctx, "INSERT INTO Links(uid, alias_id, linked_alias_id) VALUES ($1, "+
				"CASE WHEN ($2 > 0) THEN $2 ELSE (SELECT id FROM aliases WHERE name = $3) END, "+
				"(SELECT id FROM aliases WHERE name = $4))",
				uid, aliasID, name, refAlias); err == nil {
				err = tx.Commit()
			}
		}
	}

	reply := replierFactory(reqenv, msg)
	if err != nil {
		log.Error(err)
		if err := tx.Rollback(); err != nil {
			log.Error(err)
		}
		reply(LinkStatusFailure)
	} else if isDuplicateConstraintViolation(err) {
		reply(LinkStatusDuplicate)
	} else {
		reply(LinkStatusSuccess)
	}
}
