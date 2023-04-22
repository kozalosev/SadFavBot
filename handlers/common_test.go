package handlers

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"testing"
)

var (
	container testcontainers.Container
	db        *pgxpool.Pool
	ctx       = context.Background()
)

// TestMain controls main for the tests and allows for setup and shutdown of tests
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
		Name:         "SadFavBot-HandlersTest-Postgres",
		Image:        "postgres:latest",
		ExposedPorts: []string{test.ExposedDBPort + "/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		Env: map[string]string{
			"POSTGRES_USER":     test.User,
			"POSTGRES_PASSWORD": test.Password,
			"POSTGRES_DB":       test.DB,
		},
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

	host, err := container.Host(ctx)
	containerPort, err := container.MappedPort(ctx, test.ExposedDBPort)
	port := strings.TrimSuffix(string(containerPort), "/tcp")

	dbConfig := storage.NewDatabaseConfig(host, port, test.User, test.Password, test.DB)
	db = storage.ConnectToDatabase(ctx, dbConfig)
	storage.RunMigrations(dbConfig, "")
}

func shutDown() {
	if err := container.Terminate(ctx); err != nil {
		panic(fmt.Sprintf("failed to terminate container: %s", err.Error()))
	}
}
