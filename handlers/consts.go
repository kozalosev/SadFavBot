package handlers

const (
	FieldAlias     = "alias"
	FieldObject    = "object"
	FieldDeleteAll = "deleteAll"
	FieldLanguage  = "language"

	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusNoRows  = "no.rows"
	StatusDuplicate = "duplicate"

	FieldValidationErrorTrInfix = ".validation.error."
	FieldMaxLengthErrorTrSuffix = FieldValidationErrorTrInfix + "length"

	DuplicateConstraintSQLCode = "23505"
)
