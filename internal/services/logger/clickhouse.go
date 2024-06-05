package logger

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/ClickHouse/clickhouse-go"

	"websocket-confee/internal/models"
	"websocket-confee/internal/repositories/clickhouse/write"
)

type writeRepositoryInterface interface {
	BatchInsertLog(logs map[string]models.Log) error
}

type ClickhouseLogger struct {
	writeRepository writeRepositoryInterface
}

func newClickhouseLogger() (*ClickhouseLogger, error) {
	host := os.Getenv("CLICKHOUSE_HOST")
	port := os.Getenv("CLICKHOUSE_PORT")
	database := os.Getenv("CLICKHOUSE_DB")
	user := os.Getenv("CLICKHOUSE_USER")
	password := os.Getenv("CLICKHOUSE_PASSWORD")

	db, err := sql.Open(
		"clickhouse",
		fmt.Sprintf("tcp://%s:%s?username=%s&password=%s&database=%s", host, port, user, password, database),
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect to ClickHouse :%s", err))
	}

	return &ClickhouseLogger{
		writeRepository: write.NewClickhouseWriteRepository(db),
	}, nil
}

func (c *ClickhouseLogger) BatchSave(logs map[string]models.Log) {
	c.writeRepository.BatchInsertLog(logs)
}
