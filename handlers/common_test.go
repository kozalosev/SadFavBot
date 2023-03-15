package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/kozalosev/SadFavBot/storage"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
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

	TestUID           = 123456
	TestUID2          = TestUID + 1
	TestUID3          = TestUID + 2
	TestAlias         = "alias"
	TestAliasCI       = "AliAS"
	TestAliasID       = 1
	TestAlias2        = TestAlias + "2"
	TestAlias2ID      = 2
	TestType          = wizard.Sticker
	TestFileID        = "FileID"
	TestFileID2       = "FileID_2"
	TestUniqueFileID  = "FileUniqueID"
	TestUniqueFileID2 = "FileUniqueID_2"
	TestText          = "test_text"
	TestTextID        = 1
)

var (
	container testcontainers.Container
	db        *sql.DB
	ctx       = context.Background()
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

	dbConfig := storage.NewDatabaseConfig(host, port, TestUser, TestPassword, TestDB)
	db = storage.ConnectToDatabase(dbConfig)
	storage.RunMigrations(dbConfig, "")
}

func shutDown() {
	if err := container.Terminate(ctx); err != nil {
		panic(fmt.Sprintf("failed to terminate container: %s", err.Error()))
	}
}

func insertTestData(db *sql.DB) {
	for _, table := range []string{"items", "aliases", "texts", "users"} {
		_, err := db.Exec("DELETE FROM " + table)
		check(err)
	}

	_, err := db.Exec("INSERT INTO aliases(id, name) VALUES ($1, $2), ($3, $4)",
		TestAliasID, TestAlias, TestAlias2ID, TestAlias2)
	check(err)
	_, err = db.Exec("INSERT INTO items(uid, type, alias, file_id, file_unique_id) VALUES"+
		"($1, $3, $4, $6, $8),"+ // TestUID, TestAlias, TestFileID, TestUniqueFileID
		"($1, $3, $4, $7, $9),"+ // TestUID, TestAlias, TestFileID2, TestUniqueFileID2
		"($1, $3, $5, $6, $8),"+ // TestUID, TestAlias2, TestFileID, TestUniqueFileID
		"($2, $3, $4, $6, $8)", // TestUID2, TestAlias, TestFileID, TestUniqueFileID
		TestUID, TestUID2, TestType, TestAliasID, TestAlias2ID, TestFileID, TestFileID2, TestUniqueFileID, TestUniqueFileID2)
	check(err)
	_, err = db.Exec("INSERT INTO texts(id, text) VALUES ($1, $2)", TestTextID, TestText)
	check(err)
	_, err = db.Exec("INSERT INTO items(uid, type, alias, text) VALUES ($1, $2, $3, $4)",
		TestUID2, wizard.Text, TestAlias2ID, TestTextID)
	check(err)

	_, err = db.Exec("INSERT INTO users(uid, language) VALUES ($1, 'ru')", TestUID)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkRowsCount(t *testing.T, expected int, uid int64, alias *string) {
	var countRes *sql.Row
	if alias != nil {
		countRes = db.QueryRow("SELECT count(id) FROM items WHERE uid = $1 AND alias = (SELECT id FROM aliases WHERE name = $2)", uid, alias)
	} else {
		countRes = db.QueryRow("SELECT count(id) FROM items WHERE uid = $1", uid)
	}
	var count int
	err := countRes.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, expected, count)
}
