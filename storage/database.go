package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strconv"
)

func ConnectToDatabase(host, port, username, password, dbName string) *sql.DB {
	intPort, err := strconv.ParseInt(port, 10, strconv.IntSize)
	if err != nil {
		panic(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, intPort, username, password, dbName)
	conn, err := sql.Open("pgx", psqlInfo)
	if err != nil {
		panic(err)
	}

	if err := conn.Ping(); err != nil {
		panic(err)
	}
	return conn
}
