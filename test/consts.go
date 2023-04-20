package test

import "github.com/kozalosev/SadFavBot/wizard"

const (
	User          = "test"
	Password      = "testpw"
	DB            = "testdb"
	ExposedDBPort = "5432"

	UID             = 123456
	UID2            = UID + 1
	UID3            = UID + 2
	Alias           = "alias"
	AliasCI         = "AliAS"
	AliasID         = 1
	Alias2          = Alias + "'2"
	Alias2ID        = 2
	Type            = wizard.Sticker
	FileID          = "FileID"
	FileID2         = "FileID_2"
	UniqueFileID    = "FileUniqueID"
	UniqueFileID2   = "FileUniqueID_2"
	Text            = "test_text"
	TextID          = 1
	Package         = "package/test"
	PackageFullName = "123456@package/test"
	PackageID       = 1
)