package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectToDatabase(host, username, password, dbName string) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, 5432, username, password, dbName)
	conn, err := sql.Open("pgx", psqlInfo)
	if err != nil {
		panic(err)
	}
	return conn
}
