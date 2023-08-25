package repo

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
)

var sqlEscaper = strings.NewReplacer(
	"%", "\\%",
	"?", "\\?")

type RowsAffectedAware interface {
	RowsAffected() int64
}

// FavService is a common CRUD service for Favs, Aliases and Texts tables.
// It knows about the links as well.
type FavService struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewFavsService(appenv *base.ApplicationEnv) *FavService {
	return &FavService{
		ctx: appenv.Ctx,
		db:  appenv.Database,
	}
}

// Find is a method to search for favs in the database.
// The search can be performed either by exact match or by substring match.
func (service *FavService) Find(uid int64, query string, bySubstr bool) ([]*dto.Fav, error) {
	query = sqlEscaper.Replace(query)
	if bySubstr {
		query = "%" + query + "%"
	}

	q := "SELECT DISTINCT ON (file_unique_id, text_id, location_id) f.id, type, file_id, t.text, loc.latitude, loc.longitude FROM favs f " +
		"JOIN aliases a ON a.id = f.alias_id " +
		"LEFT JOIN texts t ON t.id = f.text_id " +
		"LEFT JOIN locations loc ON loc.id = f.location_id " +
		"LEFT JOIN alias_visibility av ON av.uid = f.uid AND av.alias_id = f.alias_id " +
		"WHERE f.uid = $1 AND (name ILIKE $2 OR name IN (SELECT ai_linked.name FROM links l " +
		"	JOIN aliases ai ON l.alias_id = ai.id " +
		"	JOIN aliases ai_linked ON l.linked_alias_id = ai_linked.id " +
		"	WHERE l.uid = $1 AND ai.name ILIKE $2)) " +
		"  AND ($3 IS FALSE OR av.hidden IS NOT TRUE OR lower('%' || name || '%') = lower($2)) " + // only exact match for hidden favs!
		"LIMIT 50"
	rows, err := service.db.Query(service.ctx, q, uid, query, bySubstr)

	var result []*dto.Fav
	if err != nil {
		return result, err
	}
	for rows.Next() {
		row := &dto.Fav{}
		var (
			fileID              *string
			latitude, longitude *float64
		)
		err = rows.Scan(&row.ID, &row.Type, &fileID, &row.Text, &latitude, &longitude)
		if err != nil {
			log.WithField(logconst.FieldService, "FavService").
				WithField(logconst.FieldMethod, "Find").
				WithField(logconst.FieldCalledObject, "Rows").
				WithField(logconst.FieldCalledMethod, "Scan").
				Error(err)
			continue
		}
		if fileID != nil {
			row.File = &wizard.File{ID: *fileID}
		}
		if latitude != nil {
			if longitude == nil {
				return nil, errors.New("unexpected absence of longitude when latitude is present")
			}
			row.Location = &wizard.LocData{Latitude: *latitude, Longitude: *longitude}
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// Save a fav associated with the user and alias.
func (service *FavService) Save(uid int64, alias string, fav *dto.Fav) (RowsAffectedAware, error) {
	tx, err := service.db.Begin(service.ctx)
	if err != nil {
		return nil, err
	}
	var res pgconn.CommandTag
	switch fav.Type {
	case wizard.Text:
		res, err = service.saveText(tx, uid, alias, *fav.Text)
	case wizard.Location:
		res, err = service.saveLocation(tx, uid, alias, *fav.Location)
	default:
		res, err = service.saveFile(tx, uid, alias, fav.Type, *fav.File)
	}
	if err != nil {
		if err := tx.Rollback(service.ctx); err != nil {
			log.WithField(logconst.FieldService, "FavService").
				WithField(logconst.FieldMethod, "Save").
				WithField(logconst.FieldCalledObject, "Tx").
				WithField(logconst.FieldCalledMethod, "Rollback").
				Error(err)
		}
		return &res, err
	}
	return &res, tx.Commit(service.ctx)
}

// DeleteByAlias deletes all the user's favs associated with alias.
func (service *FavService) DeleteByAlias(uid int64, alias string) (RowsAffectedAware, error) {
	log.WithField(logconst.FieldService, "FavService").
		WithField(logconst.FieldMethod, "DeleteByAlias").
		Infof("Deletion of favs and/or links with uid '%d' and alias '%s'", uid, alias)
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
			if err := tx.Rollback(service.ctx); err != nil {
				log.WithField(logconst.FieldService, "FavService").
					WithField(logconst.FieldMethod, "DeleteByAlias").
					WithField(logconst.FieldCalledObject, "Tx").
					WithField(logconst.FieldCalledMethod, "Rollback").
					Error(err)
			}
		}
	}
	return &resUnion, err
}

// DeleteFav deletes a specific fav of the user.
func (service *FavService) DeleteFav(uid int64, alias string, fav *dto.Fav) (RowsAffectedAware, error) {
	switch fav.Type {
	case wizard.Text:
		return service.deleteByText(uid, alias, *fav.Text)
	case wizard.Location:
		return service.deleteByLocation(uid, alias, *fav.Location)
	default:
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

func (service *FavService) saveLocation(tx pgx.Tx, uid int64, alias string, location wizard.LocData) (pgconn.CommandTag, error) {
	var (
		aliasID, locationID int
		err                 error
	)
	if aliasID, err = saveAliasToSeparateTable(service.ctx, tx, alias); err == nil {
		if locationID, err = saveLocationToSeparateTable(service.ctx, tx, location.Latitude, location.Longitude); err == nil {
			return tx.Exec(service.ctx, "INSERT INTO favs (uid, type, alias_id, location_id) VALUES ($1, $2, "+
				"CASE WHEN ($3 > 0) THEN $3 ELSE (SELECT id FROM aliases WHERE name = $4) END, "+
				"CASE WHEN ($5 > 0) THEN $5 ELSE (SELECT id FROM locations WHERE latitude = $6 AND longitude = $7) END)",
				uid, wizard.Location, aliasID, alias, locationID, location.Latitude, location.Longitude)
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
	log.WithField(logconst.FieldService, "FavService").
		WithField(logconst.FieldMethod, "deleteByText").
		Infof("Deletion of fav with uid '%d', alias '%s' and text '%s'", uid, alias, text)
	return service.db.Exec(service.ctx,
		"DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2) AND text_id = (SELECT id FROM texts WHERE text = $3)",
		uid, alias, text)
}

func (service *FavService) deleteByLocation(uid int64, alias string, location wizard.LocData) (RowsAffectedAware, error) {
	log.WithField(logconst.FieldService, "FavService").
		WithField(logconst.FieldMethod, "deleteByLocation").
		Infof("Deletion of fav with uid '%d', alias '%s' and location (%f, %f)", uid, alias, location.Latitude, location.Longitude)
	return service.db.Exec(service.ctx,
		"DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2) AND location_id = (SELECT id FROM locations WHERE latitude = $3 AND longitude = $4)",
		uid, alias, location.Latitude, location.Longitude)
}

func (service *FavService) deleteByFileID(uid int64, alias string, file wizard.File) (RowsAffectedAware, error) {
	log.WithField(logconst.FieldService, "FavService").
		WithField(logconst.FieldMethod, "deleteByFileID").
		Infof("Deletion of fav with uid '%d', alias '%s' and file_id '%s'", uid, alias, file.UniqueID)
	return service.db.Exec(service.ctx,
		"DELETE FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2) AND file_unique_id = $3",
		uid, alias, file.UniqueID)
}
