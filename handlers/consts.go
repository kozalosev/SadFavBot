package handlers

import "github.com/kozalosev/SadFavBot/handlers/common"

const (
	FieldAlias     = common.FieldAlias
	FieldObject    = common.FieldObject
	FieldDeleteAll = "deleteAll"
	FieldLanguage  = "language"

	StatusSuccess   = "success"
	StatusFailure   = "failure"
	StatusNoRows    = "no.rows"
	StatusDuplicate = "duplicate"

	FieldValidationErrorTrInfix = ".validation.error."
	FieldMaxLengthErrorTrSuffix = FieldValidationErrorTrInfix + "length"

	DuplicateConstraintSQLCode = "23505"
)
