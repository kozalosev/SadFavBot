package storage

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/golang-migrate/migrate/source/github"
	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

const migrationsPath = "db/migrations"

type DatabaseConfig struct {
	host string
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

func ConnectToDatabase(config *DatabaseConfig) *sql.DB {
	intPort, err := strconv.ParseInt(config.port, 10, strconv.IntSize)
	if err != nil {
		panic(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.host, intPort, config.user, config.password, config.dbName)
	conn, err := sql.Open("pgx", psqlInfo)
	if err != nil {
		panic(err)
	}

	if err := conn.Ping(); err != nil {
		panic(err)
	}
	return conn
}

func RunMigrations(config *DatabaseConfig, migrationsRepo string) {
	var sourceURL string
	if _, err := os.Stat(migrationsPath); err == nil {
		sourceURL = "file://" + migrationsPath
	} else if _, err := os.Stat("../" + migrationsPath); err == nil {
		sourceURL = "file://../" + migrationsPath
	} else {
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
