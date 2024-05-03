package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var trimCountRegex = regexp.MustCompile("\\(\\d+\\)$")

var ResultsPerPage int

func init() {
	if parsed, err := strconv.Atoi(os.Getenv("RESULTS_PER_PAGE")); err != nil {
		log.WithField(logconst.FieldFunc, "init").
			WithField(logconst.FieldConst, "RESULTS_PER_PAGE").
			Error(err)
		ResultsPerPage = 100 // default
	} else {
		ResultsPerPage = parsed
	}
}

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

type Page struct {
	Items       []string
	HasNextPage bool

	ofPackages bool
}

func (p Page) GetLastItem() string {
	last := p.Items[len(p.Items)-1]
	if p.ofPackages {
		return strings.SplitN(last, "@", 2)[1]
	} else {
		return last
	}
}

// ListWithCounts prints the list in the format:
//
//	alias (1)
//	link -> alias
//
// where '1' is the count of favs associated with 'alias'
func (service *AliasService) ListWithCounts(uid int64, grep string, lastAlias string) (*Page, error) {
	q := "SELECT a1.name, count(a1.name), null AS link FROM favs f " +
		"JOIN aliases a1 ON f.alias_id = a1.id " +
		"LEFT JOIN alias_visibility av ON av.uid = f.uid AND av.alias_id = f.alias_id " +
		"WHERE f.uid = $1 AND av.hidden IS NOT true AND ($2 = '' OR name ILIKE $2) AND name > $3 " +
		"GROUP BY a1.name " +
		"UNION " +
		"SELECT a2.name, null AS count, (SELECT name FROM aliases a WHERE a.id = l.linked_alias_id) AS link FROM links l " +
		"JOIN aliases a2 ON l.alias_id = a2.id " +
		"LEFT JOIN alias_visibility av ON av.uid = l.uid AND av.alias_id = l.alias_id " +
		"WHERE l.uid = $1 AND av.hidden IS NOT true AND ($2 = '' OR name ILIKE $2) AND name > $3 " +
		"ORDER BY name LIMIT $4"

	grep = "%" + sqlEscaper.Replace(grep) + "%"
	if rows, err := service.db.Query(service.ctx, q, uid, grep, lastAlias, ResultsPerPage+1); err == nil {
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
				log.WithField(logconst.FieldService, "AliasService").
					WithField(logconst.FieldMethod, "ListWithCounts").
					WithField(logconst.FieldCalledObject, "Rows").
					WithField(logconst.FieldCalledMethod, "Scan").
					Error(err)
			}
		}
		if len(aliases) > ResultsPerPage {
			page := &Page{
				Items:       aliases[:ResultsPerPage],
				HasNextPage: true,
			}
			return page, rows.Err()
		} else {
			page := &Page{
				Items:       aliases,
				HasNextPage: false,
			}
			return page, rows.Err()
		}
	} else {
		return nil, err
	}
}

// List returns the list of the user's aliases.
func (service *AliasService) List(uid int64) ([]string, error) {
	res, err := service.ListWithCounts(uid, "", "")
	if err == nil {
		return funk.Map(res.Items, trimSuffix).([]string), nil
	}
	return nil, err
}

// ListForFavsOnly returns the list of the user's aliases associated only with favs, but not with links.
func (service *AliasService) ListForFavsOnly(uid int64) ([]string, error) {
	if res, err := service.db.Query(service.ctx,
		"SELECT DISTINCT a.name FROM favs f "+
			"JOIN aliases a on a.id = f.alias_id "+
			"LEFT JOIN alias_visibility av ON av.uid = f.uid AND av.alias_id = f.alias_id "+
			"WHERE f.uid = $1 AND av.hidden IS NOT true "+
			"ORDER BY a.name", uid); err == nil {
		var (
			aliases []string
			alias   string
		)
		for res.Next() {
			if err := res.Scan(&alias); err == nil {
				aliases = append(aliases, alias)
			} else {
				log.WithField(logconst.FieldService, "AliasService").
					WithField(logconst.FieldMethod, "ListForFavsOnly").
					WithField(logconst.FieldCalledObject, "Rows").
					WithField(logconst.FieldCalledMethod, "Scan").
					Error(err)
			}
		}
		return aliases, res.Err()
	} else {
		return nil, err
	}
}

// ListByFile returns the list of the user's aliases associated with a specific file (i.e. any object which is not text nor location).
func (service *AliasService) ListByFile(uid int64, file *wizard.File) ([]string, error) {
	res, err := service.db.Query(service.ctx,
		"SELECT DISTINCT a.name FROM favs f "+
			"JOIN aliases a ON f.alias_id = a.id "+
			"WHERE f.uid = $1 AND f.file_unique_id = $2",
		uid, file.UniqueID)
	if err != nil {
		return nil, err
	}

	var (
		aliases []string
		alias   string
	)
	for res.Next() {
		if err := res.Scan(&alias); err == nil {
			aliases = append(aliases, alias)
		} else {
			log.WithField(logconst.FieldService, "AliasService").
				WithField(logconst.FieldMethod, "ListByFile").
				WithField(logconst.FieldCalledObject, "Rows").
				WithField(logconst.FieldCalledMethod, "Scan").
				Error(err)
		}
	}
	return aliases, res.Err()
}

// ListByText returns the list of the user's aliases associated with a specific textual fav.
func (service *AliasService) ListByText(uid int64, text *wizard.Txt) ([]string, error) {
	sql := "SELECT DISTINCT a.name FROM favs f " +
		"JOIN aliases a ON f.alias_id = a.id " +
		"JOIN texts t ON f.text_id = t.id " +
		"WHERE f.uid = $1 AND t.text = $2 AND t.entities = "

	var (
		res pgx.Rows
		err error
	)
	if len(text.Entities) > 0 {
		var entities []byte
		if entities, err = json.Marshal(text.Entities); err == nil {
			res, err = service.db.Query(service.ctx, sql+"$3", uid, text.Value, entities)
		}
	} else {
		res, err = service.db.Query(service.ctx, sql+"'null'", uid, text.Value)
	}
	if err != nil {
		return nil, err
	}

	var (
		aliases []string
		alias   string
	)
	for res.Next() {
		if err := res.Scan(&alias); err == nil {
			aliases = append(aliases, alias)
		} else {
			log.WithField(logconst.FieldService, "AliasService").
				WithField(logconst.FieldMethod, "ListByText").
				WithField(logconst.FieldCalledObject, "Rows").
				WithField(logconst.FieldCalledMethod, "Scan").
				Error(err)
		}
	}
	return aliases, res.Err()
}

// ListByLocation returns the list of the user's aliases associated with a specific location fav.
func (service *AliasService) ListByLocation(uid int64, loc *wizard.LocData) ([]string, error) {
	res, err := service.db.Query(service.ctx,
		"SELECT DISTINCT a.name FROM favs f "+
			"JOIN aliases a ON f.alias_id = a.id "+
			"JOIN locations loc ON f.location_id = loc.id "+
			"WHERE f.uid = $1 AND loc.latitude = $2 AND loc.longitude = $3",
		uid, loc.Latitude, loc.Longitude)
	if err != nil {
		return nil, err
	}

	var (
		aliases []string
		alias   string
	)
	for res.Next() {
		if err := res.Scan(&alias); err == nil {
			aliases = append(aliases, alias)
		} else {
			log.WithField(logconst.FieldService, "AliasService").
				WithField(logconst.FieldMethod, "ListByLocation").
				WithField(logconst.FieldCalledObject, "Rows").
				WithField(logconst.FieldCalledMethod, "Scan").
				Error(err)
		}
	}
	return aliases, res.Err()
}

// Hide excludes all favs associated with a specified alias from the output of List, ListWithCounts and ListForFavsOnly methods.
func (service *AliasService) Hide(uid int64, alias string) error {
	return service.changeVisibility(uid, alias, true)
}

// Reveal is a reversed to Hide operation.
func (service *AliasService) Reveal(uid int64, alias string) error {
	return service.changeVisibility(uid, alias, false)
}

func (service *AliasService) changeVisibility(uid int64, alias string, hidden bool) error {
	res, err := service.db.Exec(service.ctx,
		"INSERT INTO Alias_Visibility (uid, alias_id, hidden) VALUES ($1, (SELECT id FROM aliases WHERE name = $2), $3) ON CONFLICT (uid, alias_id) DO UPDATE SET hidden = $3",
		uid, alias, hidden)
	if err != nil {
		return err
	} else if res.RowsAffected() < 1 {
		return NoRowsWereAffected
	} else {
		return nil
	}
}

func trimCountSuffix(s string) string {
	if indexes := trimCountRegex.FindStringIndex(s); indexes != nil {
		return strings.TrimSpace(s[:indexes[0]])
	} else {
		return s
	}
}

func trimLinkSuffix(s string) string {
	if i := strings.Index(s, "→"); i >= 0 {
		return strings.TrimSpace(s[:i])
	} else {
		return s
	}
}

func trimSuffix(s string) string {
	s = trimCountSuffix(s)
	s = trimLinkSuffix(s)
	return s
}
