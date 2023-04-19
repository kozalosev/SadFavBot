package settings

// LangCode is a language code like 'ru' or 'en'.
type LangCode string

// UserOptions stored in the database.
type UserOptions struct {
	SubstrSearchEnabled bool
}

type OptionsFetcher interface {
	FetchUserOptions(uid int64, defaultLang string) (LangCode, *UserOptions)
}
