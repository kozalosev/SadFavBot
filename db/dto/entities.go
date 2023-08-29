// Package dto consists of data transfer objects or entities.
package dto

import "github.com/kozalosev/goSadTgBot/wizard"

// UserOptions stored in the database.
type UserOptions struct {
	SubstrSearchEnabled bool
}

// Fav is a favorite.
// https://github.com/kozalosev/SadFavBot/wiki/Glossary#fav
type Fav struct {
	ID       string
	Type     wizard.FieldType
	File     *wizard.File
	Text     *wizard.Txt
	Location *wizard.LocData
}
