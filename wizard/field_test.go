package wizard

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFields_FindField(t *testing.T) {
	tnField := Field{Name: TestName}
	tn2Field := Field{Name: TestName2}
	fields := Fields{&tnField, &tn2Field}

	res := fields.FindField(TestName)

	assert.Equal(t, &tnField, res)
	assert.NotEqual(t, &tn2Field, res)
}

func TestFields_FindField_MultipleItems(t *testing.T) {
	tnField := Field{Name: TestName}
	tn2Field := Field{Name: TestName}
	tn3Field := Field{Name: TestName2}
	fields := Fields{&tnField, &tn2Field, &tn3Field}

	res := fields.FindField(TestName)

	assert.Equal(t, &tnField, res)
	assert.NotSame(t, &tn2Field, res)
	assert.NotEqual(t, &tn3Field, res)
}

func TestFields_FindField_NotExistentField(t *testing.T) {
	tnField := Field{Name: TestName}
	fields := Fields{&tnField}

	res := fields.FindField(TestName2)

	assert.Nil(t, res)
}

func TestFieldMarshalling(t *testing.T) {
	field := Field{
		Name:         TestName,
		Data:         TestValue,
		WasRequested: true,
		Type:         Text,
	}

	jsonBytes, err := json.Marshal(field)
	assert.NoError(t, err)
	jsn := string(jsonBytes)

	entities := []string{
		TestName, TestValue, "true", string(Text),
	}
	for _, e := range entities {
		assert.Contains(t, jsn, e)
	}

	var restoredField Field
	err = json.Unmarshal(jsonBytes, &restoredField)
	assert.NoError(t, err)
	assert.Equal(t, field, restoredField)
}
