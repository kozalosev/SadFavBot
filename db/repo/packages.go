package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
)

var NoRowsWereAffected = errors.New("no rows were affected")

// PackageService is a CRUD service for the Package entity.
// It's also responsible for installation of packages.
// https://github.com/kozalosev/SadFavBot/wiki/Glossary#package
type PackageService struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewPackageService(appenv *base.ApplicationEnv) *PackageService {
	return &PackageService{
		ctx: appenv.Ctx,
		db:  appenv.Database,
	}
}

// ResolveName fetches the name of a package by its unique_id.
func (service *PackageService) ResolveName(uuid string) (string, error) {
	var (
		uid  int
		name string
	)
	q := "SELECT owner_uid, name FROM packages WHERE unique_id = $1"
	err := service.db.QueryRow(service.ctx, q, uuid).Scan(&uid, &name)
	return fmt.Sprintf("%d@%s", uid, name), err
}

// ListWithCounts prints the list of saved packages of some user in the format:
//
//	package (1)
//
// where '1' is the count of aliases associated with 'package'
func (service *PackageService) ListWithCounts(uid int64, lastPackage string) (*Page, error) {
	q := "SELECT p.name, count(pa.alias_id) FROM packages p " +
		"JOIN package_aliases pa ON p.id = pa.package_id " +
		"WHERE p.owner_uid = $1 AND p.name > $2 " +
		"GROUP BY p.name ORDER BY p.name LIMIT $3"

	if rows, err := service.db.Query(service.ctx, q, uid, lastPackage, ResultsPerPage+1); err == nil {
		var (
			packages []string
			pkg      string
			count    int
		)
		for rows.Next() {
			if err = rows.Scan(&pkg, &count); err == nil {
				packages = append(packages, fmt.Sprintf("%s (%d)", pkg, count))
			} else {
				log.WithField(logconst.FieldService, "PackageService").
					WithField(logconst.FieldMethod, "ListWithCounts").
					WithField(logconst.FieldCalledObject, "Rows").
					WithField(logconst.FieldCalledMethod, "Scan").
					Error(err)
			}
		}
		lines := funk.Map(packages, func(s string) string {
			return FormatPackageName(uid, s)
		}).([]string)
		if len(lines) > ResultsPerPage {
			return &Page{
				Items:       lines[:ResultsPerPage],
				HasNextPage: true,
				ofPackages:  true,
			}, rows.Err()
		} else {
			return &Page{
				Items:       lines,
				HasNextPage: false,
				ofPackages:  true,
			}, rows.Err()
		}
	} else {
		return nil, err
	}
}

// ListAliases returns the list of aliases in the package.
func (service *PackageService) ListAliases(pkgInfo *PackageInfo) (items []string, err error) {
	var res pgx.Rows
	q := "SELECT a.name FROM package_aliases pa JOIN packages p ON p.id = pa.package_id JOIN aliases a ON pa.alias_id = a.id WHERE p.owner_uid = $1 AND p.name = $2"
	if res, err = service.db.Query(service.ctx, q, pkgInfo.UID, pkgInfo.Name); err == nil {
		var item string
		for res.Next() {
			if err := res.Scan(&item); err == nil {
				items = append(items, item)
			} else {
				log.WithField(logconst.FieldService, "PackageService").
					WithField(logconst.FieldMethod, "ListAliases").
					WithField(logconst.FieldCalledObject, "Rows").
					WithField(logconst.FieldCalledMethod, "Scan").
					Error(err)
			}
		}
		err = res.Err()
	}
	return
}

// Exists returns true if the package exists in the database.
func (service *PackageService) Exists(pkgInfo *PackageInfo) (exists bool, err error) {
	q := "SELECT exists(SELECT 1 FROM Packages WHERE owner_uid = $1 AND name = $2)"
	err = service.db.QueryRow(service.ctx, q, pkgInfo.UID, pkgInfo.Name).Scan(&exists)
	return
}

// Create a new package.
func (service *PackageService) Create(uid int64, name string, aliases []string) (string, error) {
	var (
		packID string
		tx     pgx.Tx
		err    error
	)
	if tx, err = service.db.Begin(service.ctx); err == nil {
		if packID, err = service.createPackage(tx, uid, name, aliases); err == nil {
			err = tx.Commit(service.ctx)
		}
	}
	return packID, err
}

// Delete some package.
func (service *PackageService) Delete(uid int64, name string) error {
	res, err := service.db.Exec(service.ctx, "DELETE FROM packages WHERE owner_uid = $1 AND name = $2", uid, name)
	if err != nil {
		return err
	}
	if res.RowsAffected() < 1 {
		return NoRowsWereAffected
	}
	return nil
}

// Recreate is a combination of Delete and Create executing in one transaction.
func (service *PackageService) Recreate(uid int64, name string, aliases []string) (string, error) {
	var (
		packID string
		tx     pgx.Tx
		err    error
	)
	if tx, err = service.db.Begin(service.ctx); err == nil {
		if err = service.Delete(uid, name); err == nil {
			if packID, err = service.createPackage(tx, uid, name, aliases); err == nil {
				err = tx.Commit(service.ctx)
			}
		}
	}
	return packID, err
}

func (service *PackageService) createPackage(tx pgx.Tx, uid int64, name string, aliases []string) (string, error) {
	var (
		packID   int
		packUUID string
		res      pgx.Rows
		err      error
	)
	if err = tx.QueryRow(service.ctx, "INSERT INTO packages(owner_uid, name) VALUES ($1, $2) RETURNING id, unique_id", uid, name).Scan(&packID, &packUUID); err == nil {
		aliases = funk.Map(aliases, func(a string) string {
			return strings.Replace(a, "'", "''", -1)
		}).([]string)
		if res, err = service.db.Query(service.ctx, fmt.Sprintf("SELECT id FROM aliases WHERE name IN ('%s')", strings.Join(aliases, "', '"))); err == nil {
			var (
				aliasID int
				b       pgx.Batch
			)
			for res.Next() {
				if err = res.Scan(&aliasID); err == nil {
					b.Queue("INSERT INTO package_aliases(package_id, alias_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", packID, aliasID)
				} else {
					log.WithField(logconst.FieldService, "PackageService").
						WithField(logconst.FieldMethod, "createPackage").
						WithField(logconst.FieldCalledObject, "Rows").
						WithField(logconst.FieldCalledMethod, "Scan").
						Error(err)
				}
			}
			batchRes := tx.SendBatch(service.ctx, &b)
			err = batchRes.Close()
		}
	}
	return packUUID, err
}

// FormatPackageName returns the full name of the package.
func FormatPackageName(uid int64, name string) string {
	return fmt.Sprintf("%d@%s", uid, name)
}
