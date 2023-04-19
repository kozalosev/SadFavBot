package storage

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

const migrationsPath = "db/migrations"

type DatabaseConfig struct {
	host     string
	port     string
	user     string
	password string
	dbName   string
}

func NewDatabaseConfig(host, port, username, password, dbName string) *DatabaseConfig {
	return &DatabaseConfig{
		host:     host,
		port:     port,
		user:     username,
		password: password,
		dbName:   dbName,
	}
}

// ConnectToDatabase returns a connection pool, which can be used to execute queries to the database.
func ConnectToDatabase(ctx context.Context, config *DatabaseConfig) *pgxpool.Pool {
	intPort, err := strconv.ParseInt(config.port, 10, strconv.IntSize)
	if err != nil {
		log.Fatal(err)
	}

	connURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.user, config.password, config.host, intPort, config.dbName)
	conn, err := pgxpool.New(ctx, connURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.Ping(ctx); err != nil {
		panic(err)
	}
	return conn
}

// RunMigrations either from source code on a local machine if available (for developers) or from a GitHub repository (for production).
func RunMigrations(config *DatabaseConfig, migrationsRepo string) {
	var sourceURL string
	if _, err := os.Stat(migrationsPath); err == nil {
		sourceURL = "file://" + migrationsPath
	} else if _, err := os.Stat("../" + migrationsPath); err == nil {
		sourceURL = "file://../" + migrationsPath
	} else if _, err := os.Stat("../../" + migrationsPath); err == nil {
		sourceURL = "file://../../" + migrationsPath
	} else {
		log.Warning("Run migrations from the repository")
		sourceURL = "github://" + migrationsRepo + "/" + migrationsPath
	}
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.user, config.password, config.host, config.port, config.dbName)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}
