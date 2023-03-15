package handlers

import (
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteFormAction(t *testing.T) {
	insertTestData(db)

	msg := buildMessage(TestUID)
	reqenv := buildRequestEnv()
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldDeleteAll, Data: No},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: wizard.File{UniqueID: TestUniqueFileID}},
	}

	deleteFormAction(reqenv, msg, fields)

	testAlias := TestAlias
	checkRowsCount(t, 1, TestUID, &testAlias) // row with FileID_2 is on its place
	checkRowsCount(t, 2, TestUID, nil)        // rows with alias2 and alias+FileID_2

	fields.FindField(FieldDeleteAll).Data = Yes
	deleteFormAction(reqenv, msg, fields)

	checkRowsCount(t, 0, TestUID, &testAlias)
}

func TestDeleteFormActionText(t *testing.T) {
	insertTestData(db)

	msg := buildMessage(TestUID3)
	reqenv := buildRequestEnv()
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: TestAlias},
		&wizard.Field{Name: FieldDeleteAll, Data: No},
		&wizard.Field{Name: FieldObject, Type: TestType, Data: TestText},
	}

	deleteFormAction(reqenv, msg, fields)

	checkRowsCount(t, 0, TestUID3, nil) // row with TestFileID is on its place
}

func TestTrimCountSuffix(t *testing.T) {
	assert.Equal(t, TestAlias, trimCountSuffix(TestAlias + " (1)"))
	assert.Equal(t, TestAlias, trimCountSuffix(TestAlias))
	assert.Equal(t, TestAlias + " (test)", trimCountSuffix(TestAlias + " (test)"))
}
