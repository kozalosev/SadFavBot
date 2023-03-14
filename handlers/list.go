package handlers

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	ListStatusTrPrefix = "commands.list.status."
	ListStatusSuccess  = ListStatusTrPrefix + StatusSuccess
	ListStatusFailure  = ListStatusTrPrefix + StatusFailure
	ListStatusNoRows   = ListStatusTrPrefix + StatusNoRows
)

type ListHandler struct{}

func (ListHandler) CanHandle(msg *tgbotapi.Message) bool {
	return msg.Command() == "list"
}

func (handler ListHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {
	aliases, err := fetchAliases(reqenv.Database, msg.From.ID)
	replyWith := replierFactory(reqenv, msg)
	if err != nil {
		log.Errorln(err)
		replyWith(ListStatusFailure)
	} else if len(aliases) == 0 {
		replyWith(ListStatusNoRows)
	} else {
		title := reqenv.Lang.Tr(ListStatusSuccess)
		reqenv.Bot.Reply(msg, title+"\n\n• "+strings.Join(aliases, "\n• "))
	}
}

func fetchAliases(db *sql.DB, uid int64) ([]string, error) {
	rows, err := db.Query("SELECT a.name, count(a.name) FROM items i JOIN aliases a ON i.alias = a.id WHERE uid = $1 GROUP BY a.name ORDER BY a.name", uid)
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Error(err)
		}
	}(rows)
	if err != nil {
		return nil, err
	}

	var aliases []string
	for rows.Next() {
		var (
			alias string
			count int
		)
		err = rows.Scan(&alias, &count)
		if err != nil {
			log.Error("Error occurred while fetching from database: ", err)
			continue
		}
		aliases = append(aliases, fmt.Sprintf("%s (%d)", alias, count))
	}
	return aliases, nil
}
