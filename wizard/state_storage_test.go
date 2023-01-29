package wizard

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"reflect"
	"testing"
)

var formExample = Form{
	Fields: Fields{&Field{
		Name:              "test",
		Data:              "test",
		WasRequested:      true,
		Type:              Text,
		PromptDescription: "test",
	}},
	WizardType: "TestWizard",
}

var container testcontainers.Container

func TestConnectToRedis(t *testing.T) {
	stateStorage := buildStateStorage(t)
	if err := stateStorage.Close(); err != nil {
		t.Error(err)
	}
}

func TestRedisStateStorage_SaveState(t *testing.T) {
	stateStorage := buildStateStorage(t)
	defer func() {
		if err := stateStorage.Close(); err != nil {
			t.Error(err)
		}
	}()

	copyOfForm := formExample
	err := stateStorage.SaveState(TestID, &copyOfForm)
	if err != nil {
		t.Error(err)
	}
}

func TestRedisStateStorage_GetCurrentState(t *testing.T) {
	stateStorage := buildStateStorage(t)
	defer func() {
		if err := stateStorage.Close(); err != nil {
			t.Error(err)
		}
	}()

	var f Form
	if err := stateStorage.GetCurrentState(TestID, &f); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(f, formExample) {
		t.Error("Forms are not equal!", f, formExample)
	}
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
	container, err = testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		panic(err)
	}
}

func shutDown() {
	if err := container.Terminate(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to terminate container: %s", err.Error()))
	}
}

func buildStateStorage(t *testing.T) StateStorage {
	endpoint, err := container.Endpoint(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	return ConnectToRedis(&redis.Options{Addr: endpoint})
}
