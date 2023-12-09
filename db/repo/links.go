package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	log "github.com/sirupsen/logrus"
)

// LinkService is a simple "create" service for the Link entity.
// https://github.com/kozalosev/SadFavBot/wiki/Glossary#link
type LinkService struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewLinkService(appenv *base.ApplicationEnv) *LinkService {
	return &LinkService{
		ctx: appenv.Ctx,
		db:  appenv.Database,
	}
}

// Create a link to refAlias for the user with the specified UID.
func (service *LinkService) Create(uid int64, name, refAlias string) error {
	var (
		tx  pgx.Tx
		err error
	)
	if tx, err = service.db.Begin(service.ctx); err == nil {
		var aliasID int
		if aliasID, err = saveAliasToSeparateTable(service.ctx, tx, name); err == nil {
			if _, err = tx.Exec(service.ctx, "INSERT INTO Links(uid, alias_id, linked_alias_id) VALUES ($1, "+
				"CASE WHEN ($2 > 0) THEN $2 ELSE (SELECT id FROM aliases WHERE name = $3) END, "+
				"(SELECT id FROM aliases WHERE name = $4))",
				uid, aliasID, name, refAlias); err == nil {
				err = tx.Commit(service.ctx)
			}
		}
	}

	if err != nil {
		if err := tx.Rollback(service.ctx); err != nil {
			log.WithField(logconst.FieldService, "LinkService").
				WithField(logconst.FieldMethod, "Create").
				WithField(logconst.FieldCalledObject, "Tx").
				WithField(logconst.FieldCalledMethod, "Rollback").
				Error(err)
		}
	}
	return err
}

func (service *LinkService) Delete(uid int64, name string) error {
	tag, err := service.db.Exec(service.ctx,
		"DELETE FROM Links WHERE uid = $1 AND alias_id = (SELECT id FROM Aliases WHERE name = $2)",
		uid, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() < 1 {
		return NoRowsWereAffected
	}
	return nil
}
