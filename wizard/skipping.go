package wizard

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
)

type SkipCondition interface {
	ShouldBeSkipped(form *Form) bool
}

type SkipOnFieldValue struct {
	Name  string
	Value string
}

func (s SkipOnFieldValue) ShouldBeSkipped(form *Form) bool {
	f := form.Fields.FindField(s.Name)
	if f == nil {
		log.Warningf("Field '%s' was not found to check if '%s' should be skipped!", s.Name, form.Fields[form.Index].Name)
		return false
	}
	return f.Data == s.Value
}

type NoSkip struct{}

func (NoSkip) ShouldBeSkipped(*Form) bool { return false }

type SkipConditionContainer struct {
	Data string
	Type string
}

func WrapCondition(c SkipCondition) (*SkipConditionContainer, error) {
	var (
		data     []byte
		condType string
		err      error
	)
	switch impl := c.(type) {
	case *SkipOnFieldValue:
		data, err = json.Marshal(*impl)
		condType = "SkipOnFieldValue"
	default:
		return nil, errors.New("only references to implementations of SkipCondition are allowed to be passed to the WrapCondition method")
	}
	if err != nil {
		return nil, err
	}
	return &SkipConditionContainer{
		Data: string(data),
		Type: condType,
	}, nil
}

func UnwrapCondition(cc *SkipConditionContainer) (SkipCondition, error) {
	if cc == nil {
		return NoSkip{}, nil
	}

	var (
		c   SkipCondition
		err error
	)
	switch cc.Type {
	case "SkipOnFieldValue":
		var cv SkipOnFieldValue
		err = json.Unmarshal([]byte(cc.Data), &cv)
		c = &cv
	default:
		return nil, errors.New("unknown type of SkipCondition")
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}
