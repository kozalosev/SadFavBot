package wizard

import (
	"github.com/kozalosev/SadFavBot/base"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPopulationOfWizardDescriptors(t *testing.T) {
	clearRegisteredDescriptors()
	assert.Len(t, registeredWizardDescriptors, 0)

	ok := PopulateWizardDescriptors([]base.MessageHandler{testHandler{}})
	assert.True(t, ok)

	desc := registeredWizardDescriptors[TestWizardName]
	assert.Equal(t, getFuncPtr(tAction), getFuncPtr(desc.action))

	assert.Len(t, registeredWizardDescriptors, 1)
	ok = PopulateWizardDescriptors([]base.MessageHandler{testHandler2{}}) // doesn't add anything if the map is not empty
	assert.False(t, ok)
	assert.Len(t, registeredWizardDescriptors, 1)
}

func TestFinders(t *testing.T) {
	formDesc := NewWizardDescriptor(tAction)
	f1Desc := formDesc.AddField(TestName, TestPromptDesc)
	formDesc.AddField(TestName2, TestPromptDesc)
	registeredWizardDescriptors[TestWizardName] = formDesc

	assert.Equal(t, findFormDescriptor(TestWizardName), formDesc)
	assert.Equal(t, f1Desc, formDesc.findFieldDescriptor(TestName))
	assert.Equal(t, TestPromptDesc, f1Desc.promptDescription)
}

func clearRegisteredDescriptors() {
	registeredWizardDescriptors = make(map[string]*FormDescriptor)
}
