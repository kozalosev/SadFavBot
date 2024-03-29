package test

import "github.com/kozalosev/goSadTgBot/wizard"

const (
	User          = "test"
	Password      = "testpw"
	DB            = "testdb"
	ExposedDBPort = "5432"

	UID             = 123456
	UID2            = UID + 1
	UID3            = UID + 2
	UIDPhoto        = UID + 3
	Alias           = "alias"
	AliasCI         = "AliAS"
	AliasID         = 1
	Alias2          = Alias + "'2"
	Alias2ID        = 2
	AliasPhoto      = "photo"
	AliasPhotoID    = 3
	Type            = wizard.Sticker
	FileID          = "FileID"
	FileID2         = "FileID_2"
	FileIDPhoto     = "FileID_Photo"
	UniqueFileID    = "FileUniqueID"
	UniqueFileID2   = "FileUniqueID_2"
	Text            = "test_text"
	TextID          = 1
	CaptionPhoto    = AliasPhoto
	CaptionPhotoID  = 2
	Latitude        = 1.1
	Longitude       = 2.2
	Package         = "package/test"
	PackageFullName = "123456@package/test"
	PackageID       = 1
	PackageUUID     = "a97a8b56-d461-47b5-a740-651bf14c501f"
)
