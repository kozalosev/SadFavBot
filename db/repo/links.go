package repo

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/base"
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
		if rbErr := tx.Rollback(service.ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
	}
	return err
}
