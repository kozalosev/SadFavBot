package repo

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
)

var sqlEscaper = strings.NewReplacer(
	"%", "\\%",
	"?", "\\?")

type RowsAffectedAware interface {
	RowsAffected() int64
}

// Fav is a favorite.
// https://github.com/kozalosev/SadFavBot/wiki/Glossary#fav
type Fav struct {
	ID   string
	Type wizard.FieldType
	File *wizard.File
	Text *string
}

func NewFav() *Fav {
	return &Fav{File: &wizard.File{}}
}

// FavService is a common CRUD service for Favs, Aliases and Texts tables.
// It knows about the links as well.
type FavService struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewFavsService(reqenv *base.RequestEnv) *FavService {
	return &FavService{
		ctx: reqenv.Ctx,
		db:  reqenv.Database,
	}
}

// Find is a method to search for favs in the database.
// The search can be performed either by exact match or by substring match.
func (service *FavService) Find(uid int64, query string, bySubstr bool) ([]*Fav, error) {
	query = sqlEscaper.Replace(query)
	if bySubstr {
		query = "%" + query + "%"
	}

	q := "SELECT min(f.id), type, file_id, t.text FROM favs f " +
		"JOIN aliases a ON a.id = f.alias_id " +
		"LEFT JOIN texts t ON t.id = f.text_id " +
		"WHERE uid = $1 AND (name ILIKE $2 OR name = (SELECT ai_linked.name FROM links l " +
		"	JOIN aliases ai ON l.alias_id = ai.id " +
		"	JOIN aliases ai_linked ON l.linked_alias_id = ai_linked.id " +
		"	WHERE l.uid = $1 AND ai.name ILIKE $2)) " +
		"GROUP BY type, file_id, t.text " +
		"LIMIT 50"
	rows, err := service.db.Query(service.ctx, q, uid, query)

	var result []*Fav
	if err != nil {
		log.Error("error occurred: ", err)
		return result, nil
	}
	for rows.Next() {
		row := NewFav()
		err = rows.Scan(&row.ID, &row.Type, &row.File.ID, &row.Text)
		if err != nil {
			log.Error("Error occurred while fetching from database: ", err)
			continue
		}
		result = append(result, row)
	}
	return result, nil
}

// Save a fav associated with the user and alias.
func (service *FavService) Save(uid int64, alias string, fav *Fav) (RowsAffectedAware, error) {
	tx, err := service.db.Begin(service.ctx)
	if err != nil {
		return nil, err
	}
	var res pgconn.CommandTag
	if fav.Type == wizard.Text {
		res, err = service.saveText(tx, uid, alias, *fav.Text)
	} else {
		res, err = service.saveFile(tx, uid, alias, fav.Type, *fav.File)
	}
	return &res, tx.Commit(service.ctx)
}

// DeleteByAlias deletes all of the user's favs associated with alias.
func (service *FavService) DeleteByAlias(uid int64, alias string) (RowsAffectedAware, error) {
	log.Infof("Deletion of favs and/or links with uid '%d' and alias '%s'", uid, alias)
	var (
		tx       pgx.Tx
		res      pgconn.CommandTag
		resUnion rowsAffectedAdder
		err      error
	)
	if tx, err = service.db.Begin(service.ctx); err == nil {
		if res, err = tx.Exec(service.ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2)", uid, alias); err == nil {
			resUnion.Add(res)
			if res, err = tx.Exec(service.ctx, "DELETE FROM links WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2)", uid, alias); err == nil {
				resUnion.Add(res)
				err = tx.Commit(service.ctx)
			}
		}
		if err != nil {
			if rbErr := tx.Rollback(service.ctx); rbErr != nil {
				err = errors.Join(err, rbErr)
			}
		}
	}
	return &resUnion, err
}

// DeleteFav deletes a specific fav of the user.
func (service *FavService) DeleteFav(uid int64, alias string, fav *Fav) (RowsAffectedAware, error) {
	if fav.Type == wizard.Text {
		return service.deleteByText(uid, alias, *fav.Text)
	} else {
		return service.deleteByFileID(uid, alias, *fav.File)
	}
}

func (service *FavService) saveText(tx pgx.Tx, uid int64, alias, text string) (pgconn.CommandTag, error) {
	var (
		aliasID, textID int
		err             error
	)
	if aliasID, err = saveAliasToSeparateTable(service.ctx, tx, alias); err == nil {
		if textID, err = saveTextToSeparateTable(service.ctx, tx, text); err == nil {
			return tx.Exec(service.ctx, "INSERT INTO favs (uid, type, alias_id, text_id) VALUES ($1, $2, "+
				"CASE WHEN ($3 > 0) THEN $3 ELSE (SELECT id FROM aliases WHERE name = $4) END, "+
				"CASE WHEN ($5 > 0) THEN $5 ELSE (SELECT id FROM texts WHERE text = $6) END)",
				uid, wizard.Text, aliasID, alias, textID, text)
		}
	}
	return pgconn.CommandTag{}, err
}

func (service *FavService) saveFile(tx pgx.Tx, uid int64, alias string, fileType wizard.FieldType, file wizard.File) (pgconn.CommandTag, error) {
	if aliasID, err := saveAliasToSeparateTable(service.ctx, tx, alias); err == nil {
		return tx.Exec(service.ctx, "INSERT INTO favs (uid, type, alias_id, file_id, file_unique_id) VALUES ($1, $2, CASE WHEN ($3 > 0) THEN $3 ELSE (SELECT id FROM aliases WHERE name = $4) END, $5, $6)",
			uid, fileType, aliasID, alias, file.ID, file.UniqueID)
	} else {
		return pgconn.CommandTag{}, err
	}

}

func (service *FavService) deleteByText(uid int64, alias, text string) (RowsAffectedAware, error) {
	log.Infof("Deletion of fav with uid '%d', alias '%s' and text '%s'", uid, alias, text)
	return service.db.Exec(service.ctx,
		"DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2) AND text_id = (SELECT id FROM texts WHERE text = $3)",
		uid, alias, text)
}

func (service *FavService) deleteByFileID(uid int64, alias string, file wizard.File) (RowsAffectedAware, error) {
	log.Infof("Deletion of fav with uid '%d', alias '%s' and file_id '%s'", uid, alias, file.UniqueID)
	return service.db.Exec(service.ctx,
		"DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2) AND file_unique_id = $3",
		uid, alias, file.UniqueID)
}
