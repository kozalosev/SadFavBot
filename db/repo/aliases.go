package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/base"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"regexp"
	"strings"
)

var trimCountRegex = regexp.MustCompile("\\(\\d+\\)$")

// AliasService is like an extension of [FavService] but to list all saved aliases of a user.
// It works with Favs and Aliases internally but returns only the latter as strings.
type AliasService struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewAliasService(appenv *base.ApplicationEnv) *AliasService {
	return &AliasService{
		ctx: appenv.Ctx,
		db:  appenv.Database,
	}
}

// ListWithCounts prints the list in the format:
//
//	alias (1)
//	link -> alias
//
// where '1' is the count of favs associated with 'alias'
func (service *AliasService) ListWithCounts(uid int64) ([]string, error) {
	q := "SELECT a1.name, count(a1.name), null AS link FROM favs f JOIN aliases a1 ON f.alias_id = a1.id WHERE f.uid = $1 GROUP BY a1.name " +
		"UNION " +
		"SELECT a2.name, null AS count, (SELECT name FROM aliases a WHERE a.id = l.linked_alias_id) AS link FROM links l JOIN aliases a2 ON l.alias_id = a2.id WHERE l.uid = $2 " +
		"ORDER BY name"

	if rows, err := service.db.Query(service.ctx, q, uid, uid); err == nil {
		var (
			aliases []string
			alias   string
			count   *int
			link    *string
		)
		for rows.Next() {
			if err = rows.Scan(&alias, &count, &link); err == nil {
				if link != nil {
					aliases = append(aliases, fmt.Sprintf("%s → %s", alias, *link))
				} else {
					aliases = append(aliases, fmt.Sprintf("%s (%d)", alias, *count))
				}
			} else {
				log.Error("Error occurred while fetching from database: ", err)
			}
		}
		return aliases, nil
	} else {
		return nil, err
	}
}

// List returns the list of the user's aliases.
func (service *AliasService) List(uid int64) ([]string, error) {
	res, err := service.ListWithCounts(uid)
	if err == nil {
		res = funk.Map(res, trimCountSuffix).([]string)
	}
	return res, err
}

// ListForFavsOnly returns the list of the user's aliases associated only with favs, but not with links.
func (service *AliasService) ListForFavsOnly(uid int64) ([]string, error) {
	if res, err := service.db.Query(service.ctx, "SELECT DISTINCT a.name FROM favs f JOIN aliases a on a.id = f.alias_id WHERE f.uid = $1", uid); err == nil {
		var (
			aliases []string
			alias   string
		)
		for res.Next() {
			if e := res.Scan(&alias); err == nil {
				aliases = append(aliases, alias)
			} else {
				err = errors.Join(err, e)
			}
		}
		return aliases, err
	} else {
		return nil, err
	}
}

func trimCountSuffix(s string) string {
	if indexes := trimCountRegex.FindStringIndex(s); indexes != nil {
		return strings.TrimSpace(s[:indexes[0]])
	} else {
		return s
	}
}