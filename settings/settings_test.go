package settings

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/kozalosev/SadFavBot/storage"
	log "github.com/sirupsen/logrus"
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

	TestUID = 123456
)

var (
	container testcontainers.Container
	db        *sql.DB
	ctx       = context.Background()
)

func TestFetchLanguage(t *testing.T) {
	clearDatabase(t)

	lang, _ := FetchUserOptions(ctx, db, TestUID, "en")
	assert.Equal(t, LangCode("en"), lang)

	res, err := db.Exec("INSERT INTO users(uid, language) VALUES ($1, 'ru')", TestUID)
	assert.NoError(t, err)
	assert.True(t, checkRowsWereAffected(res))

	lang, _ = FetchUserOptions(ctx, db, TestUID, "en")
	assert.Equal(t, LangCode("ru"), lang)
}

func TestFetchUserOptions(t *testing.T) {
	clearDatabase(t)

	_, opts := FetchUserOptions(ctx, db, TestUID, "")
	assert.False(t, opts.SubstrSearchEnabled)

	res, err := db.Exec("INSERT INTO users(uid, substring_search) VALUES ($1, true)", TestUID)
	assert.NoError(t, err)
	assert.True(t, checkRowsWereAffected(res))

	_, opts = FetchUserOptions(ctx, db, TestUID, "")
	assert.True(t, opts.SubstrSearchEnabled)
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
		Name:         "SadFavBot-MainTest-Postgres",
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

func clearDatabase(t *testing.T) {
	//goland:noinspection SqlWithoutWhere
	_, err := db.Exec("DELETE FROM users")
	assert.NoError(t, err)
}

func checkRowsWereAffected(res sql.Result) bool {
	var (
		rowsAffected int64
		err          error
	)
	if rowsAffected, err = res.RowsAffected(); err != nil {
		log.Errorln(err)
		rowsAffected = -1 // logs but ignores
	}
	if rowsAffected == 0 {
		return false
	} else {
		return true
	}
}