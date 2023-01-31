package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/kozalosev/SadFavBot/storage"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"testing"
)

const (
	TestUser      = "test"
	TestPassword  = "testpw"
	TestDB        = "testdb"
	ExposedDBPort = "5432"

	TestUID     = 123456
	TestUID2    = TestUID + 1
	TestUID3    = TestUID + 2
	TestAlias   = "alias"
	TestType    = "sticker"
	TestFileID  = "FileID"
	TestFileID2 = "FileID_2"
)

var (
	container testcontainers.Container
	dbConn    *sql.DB
)

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
		Name:         "SadFavBot-HandlersTest-Postgres",
		Image:        "postgres:latest",
		ExposedPorts: []string{ExposedDBPort + "/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		Env: map[string]string{
			"POSTGRES_USER":     TestUser,
			"POSTGRES_PASSWORD": TestPassword,
			"POSTGRES_DB":       TestDB,
		},
	}
	ctx := context.Background()
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
	containerPort, err := container.MappedPort(ctx, ExposedDBPort)
	port := strings.TrimSuffix(string(containerPort), "/tcp")

	dbConn = storage.ConnectToDatabase(host, port, TestUser, TestPassword, TestDB)
	insertTestData(dbConn)
}

func shutDown() {
	if err := container.Terminate(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to terminate container: %s", err.Error()))
	}
}

func insertTestData(dbConn *sql.DB) {
	schemaFile, err := os.ReadFile("../db/0001_schema.sql")
	check(err)
	_, err = dbConn.Exec(string(schemaFile))
	check(err)

	_, err = dbConn.Exec("INSERT INTO item(uid, type, alias, file_id) VALUES"+
		"($1, $3, $4, $6),"+ // TestUID, TestAlias, TestFileID
		"($1, $3, $4, $7),"+ // TestUID, TestAlias, TestFileID2
		"($1, $3, $5, $6),"+ // TestUID, TestAlias2, TestFileID
		"($2, $3, $4, $6)", // TestUID2, TestAlias, TestFileID
		TestUID, TestUID2, TestType, TestAlias, TestAlias+"2", TestFileID, TestFileID2)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
