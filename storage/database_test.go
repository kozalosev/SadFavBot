package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"strings"
	"testing"
)

const (
	TestUser      = "test"
	TestPassword  = "testpw"
	TestDB        = "testdb"
	ExposedDBPort = "5432"
)

func TestConnectToDatabase(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Name:         "SadFavBot-StorageTest-Postgres",
		Image:        "postgres:latest",
		ExposedPorts: []string{ExposedDBPort + "/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		Env: map[string]string{
			"POSTGRES_USER":     TestUser,
			"POSTGRES_PASSWORD": TestPassword,
			"POSTGRES_DB":       TestDB,
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	host, err := container.Host(ctx)
	assert.Nil(t, err)
	containerPort, err := container.MappedPort(ctx, ExposedDBPort)
	assert.Nil(t, err)
	port := strings.TrimSuffix(string(containerPort), "/tcp")

	dbConfig := NewDatabaseConfig(host, port, TestUser, TestPassword, TestDB)
	db := ConnectToDatabase(dbConfig)
	assert.Nil(t, db.Close())
}
