package write

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"

	"websocket-confee/internal/models"
)

type ClickhouseWriteRepository struct {
	db *sql.DB
}

func NewClickhouseWriteRepository(db *sql.DB) *ClickhouseWriteRepository {
	return &ClickhouseWriteRepository{
		db: db,
	}
}

func (c *ClickhouseWriteRepository) Log(logLevel, msg string) error {
	sqlBuilder := squirrel.Insert("logs").Values(logLevel, msg, time.Now())

	_, err := sqlBuilder.Exec()
	if err != nil {
		return err
	}

	return nil
}

func (c *ClickhouseWriteRepository) BatchInsertLog(logs map[string]models.Log) error {
	sqlBuilder := squirrel.Insert("logs")

	for key, value := range logs {
		sqlBuilder = sqlBuilder.Values(key, value.Msg, value.Time)
	}

	_, err := sqlBuilder.Exec()
	if err != nil {
		return err
	}

	return nil
}
