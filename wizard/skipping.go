package wizard

import (
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
