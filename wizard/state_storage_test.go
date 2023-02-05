package wizard

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"testing"
)

var formExample = Form{
	Fields: Fields{&Field{
		Name:         "test",
		Data:         "test",
		WasRequested: true,
		Type:         Text,
	}},
	WizardType: "TestWizard",
}

var container testcontainers.Container
var ctx = context.Background()

func TestConnectToRedis(t *testing.T) {
	stateStorage := buildStateStorage(t)
	assert.NoError(t, stateStorage.Close())
}

func TestRedisStateStorage_SaveState(t *testing.T) {
	stateStorage := buildStateStorage(t)
	defer func() {
		assert.NoError(t, stateStorage.Close())
	}()

	copyOfForm := formExample
	assert.NoError(t, stateStorage.SaveState(TestID, &copyOfForm))
}

func TestRedisStateStorage_GetCurrentState(t *testing.T) {
	stateStorage := buildStateStorage(t)
	defer func() {
		assert.NoError(t, stateStorage.Close())
	}()

	var f Form
	assert.NoError(t, stateStorage.GetCurrentState(TestID, &f))
	assert.Equal(t, formExample, f)
}

func TestRedisStateStorage_DeleteState(t *testing.T) {
	stateStorage := buildStateStorage(t)
	defer func() {
		assert.NoError(t, stateStorage.Close())
	}()

	assert.NoError(t, stateStorage.DeleteState(TestID))
}

//TestMain controls main for the tests and allows for setup and shutdown of tests
func TestMain(m *testing.M) {
	//Catching all panics to once again make sure that shutDown is successfully run
	defer func() {
		if r := recover(); r != nil {
			shutDown()
			fmt.Println("Panic", r)
		}
	}()
	setup()
	code := m.Run()
	shutDown()
	os.Exit(code)
}

func setup() {
	req := testcontainers.ContainerRequest{
		Name:         "SadFavBot-Redis",
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	var err error
	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		panic(err)
	}
}

func shutDown() {
	if err := container.Terminate(ctx); err != nil {
		panic(fmt.Sprintf("failed to terminate container: %s", err.Error()))
	}
}

func buildStateStorage(t *testing.T) StateStorage {
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	return ConnectToRedis(ctx, TestTTL, &redis.Options{Addr: endpoint})
}
