package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/settings"
	log "github.com/sirupsen/logrus"
)

// UserService is a service for working with user settings.
type UserService struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewUserService(appenv *base.ApplicationEnv) *UserService {
	return &UserService{
		ctx: appenv.Ctx,
		db:  appenv.Database,
	}
}

// Create a new user as a row in the database.
// Returns false if the user was already saved there.
func (service *UserService) Create(uid int64) (bool, error) {
	res, err := service.db.Exec(service.ctx, "INSERT INTO Users(uid) VALUES ($1) ON CONFLICT DO NOTHING", uid)
	return res.RowsAffected() > 0, err
}

// FetchUserOptions from the database if they exist.
func (service *UserService) FetchUserOptions(uid int64, defaultLang string) (settings.LangCode, *settings.UserOptions) {
	var (
		lang *string
		opts settings.UserOptions
	)
	if err := service.db.QueryRow(service.ctx, "SELECT language, substring_search FROM users WHERE uid = $1", uid).Scan(&lang, &opts.SubstrSearchEnabled); err != nil {
		log.Error(err)
		return settings.LangCode(defaultLang), &opts
	} else if lang != nil && len(*lang) > 0 {
		return settings.LangCode(*lang), &opts
	} else {
		return settings.LangCode(defaultLang), &opts
	}
}

// ChangeLanguage updates the value of the user's language option.
func (service *UserService) ChangeLanguage(uid int64, lang settings.LangCode) error {
	_, err := service.db.Exec(service.ctx, "UPDATE users SET language = $1 WHERE uid = $2", lang, uid)
	return err
}

// ChangeSubstringMode updates the value of the user's substring_mode option.
func (service *UserService) ChangeSubstringMode(uid int64, value bool) error {
	_, err := service.db.Exec(service.ctx, "UPDATE Users SET substring_search = $2 WHERE uid = $1", uid, value)
	return err
}
