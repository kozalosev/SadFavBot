package repo

import (
	log "log/slog"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/thoas/go-funk"
)

// PackageInfo is a simple struct which consists of the UID of the package's owner and the Name of the package.
type PackageInfo struct {
	UID  int64
	Name string
}

// Install the package for a specific user.
func (service *PackageService) Install(uid int64, pkgInfo *PackageInfo) ([]string, error) {
	var (
		tx           pgx.Tx
		res          pgx.Rows
		err          error
		items, links []int
	)
	if tx, err = service.db.Begin(service.ctx); err == nil {
		if items, err = service.installItems(tx, uid, pkgInfo); err == nil {
			if links, err = service.installLinks(tx, uid, pkgInfo); err == nil {
				err = tx.Commit(service.ctx)
			}
		}
	}
	aliasIDs := append(items, links...)

	if err != nil {
		if err := tx.Rollback(service.ctx); err != nil {
			log.Error("Failed to rollback a transaction",
				logconst.FieldService, "PackageService",
				logconst.FieldMethod, "Install",
				logconst.FieldCalledObject, "Tx",
				logconst.FieldCalledMethod, "Rollback",
				logconst.FieldError, err)
		}
		return nil, err
	} else if len(aliasIDs) == 0 {
		return nil, NoRowsWereAffected
	}

	aliasIDs = removeDuplicates(aliasIDs)
	aliasIDsAsStr := funk.Reduce(aliasIDs[1:], func(acc string, elem int) string {
		return acc + "," + strconv.Itoa(elem)
	}, strconv.Itoa(aliasIDs[0])).(string)

	res, err = service.db.Query(service.ctx, "SELECT name FROM aliases WHERE id IN ("+aliasIDsAsStr+")")

	var installedAliases []string
	if err == nil {
		var installedAlias string
		for res.Next() {
			if err = res.Scan(&installedAlias); err == nil {
				installedAliases = append(installedAliases, installedAlias)
			} else {
				logconst.LogFailToScanDatabaseRow("PackageService", "Install", err)
			}
		}
		err = res.Err()
	}
	return installedAliases, err
}

func (service *PackageService) installItems(tx pgx.Tx, uid int64, pkgInfo *PackageInfo) ([]int, error) {
	res, err := tx.Query(service.ctx, "INSERT INTO favs(uid, type, alias_id, file_id, file_unique_id, text_id, location_id) "+
		"SELECT cast($1 AS bigint), f.type, f.alias_id, f.file_id, f.file_unique_id, f.text_id, f.location_id FROM packages p "+
		"JOIN package_aliases pa ON p.id = pa.package_id "+
		"JOIN favs f ON f.uid = p.owner_uid AND f.alias_id = pa.alias_id "+
		"WHERE p.owner_uid = $2 AND p.name = $3 "+
		"UNION "+
		"SELECT cast($1 AS bigint), f.type, f.alias_id, f.file_id, f.file_unique_id, f.text_id, f.location_id FROM packages p "+
		"JOIN package_aliases pa ON p.id = pa.package_id "+
		"JOIN links l ON l.uid = p.owner_uid AND l.alias_id = pa.alias_id "+
		"JOIN favs f ON f.uid = p.owner_uid AND f.alias_id = l.linked_alias_id "+
		"WHERE p.owner_uid = $2 AND p.name = $3 "+
		"ON CONFLICT DO NOTHING "+
		"RETURNING alias_id", uid, pkgInfo.UID, pkgInfo.Name)
	if err == nil {
		var (
			aliasID  int
			aliasIDs []int
		)
		for res.Next() {
			if err = res.Scan(&aliasID); err == nil {
				aliasIDs = append(aliasIDs, aliasID)
			} else {
				logconst.LogFailToScanDatabaseRow("PackageService", "installItems", err)
			}
		}
		return aliasIDs, res.Err()
	}
	return nil, err
}

func (service *PackageService) installLinks(tx pgx.Tx, uid int64, pkgInfo *PackageInfo) ([]int, error) {
	res, err := tx.Query(service.ctx, "INSERT INTO links(uid, alias_id, linked_alias_id) "+
		"SELECT $1, l.alias_id, l.linked_alias_id FROM packages p "+
		"JOIN package_aliases pa ON p.id = pa.package_id "+
		"JOIN links l ON l.uid = p.owner_uid AND l.alias_id = pa.alias_id "+
		"WHERE p.owner_uid = $2 AND p.name = $3 "+
		"ON CONFLICT DO NOTHING "+
		"RETURNING alias_id", uid, pkgInfo.UID, pkgInfo.Name)
	if err == nil {
		var (
			aliasID  int
			aliasIDs []int
		)
		for res.Next() {
			if err = res.Scan(&aliasID); err == nil {
				aliasIDs = append(aliasIDs, aliasID)
			} else {
				logconst.LogFailToScanDatabaseRow("PackageService", "installLinks", err)
			}
		}
		return aliasIDs, res.Err()
	}
	return nil, err
}

func removeDuplicates[T comparable](arr []T) []T {
	set := make(map[T]struct{}, len(arr))
	for _, val := range arr {
		set[val] = struct{}{}
	}
	arr = make([]T, 0, len(set))
	for val := range set {
		arr = append(arr, val)
	}
	return arr
}
