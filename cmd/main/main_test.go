package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/redis/go-redis/v9"
)

var (
	rClient *redis.Client
	ch      *sql.DB
)

func setUp() {
	numberDb, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	rClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL") + ":" + os.Getenv("REDIS_HOST"),
		DB:   numberDb,
	})

	host := os.Getenv("CLICKHOUSE_HOST")
	port := os.Getenv("CLICKHOUSE_PORT")
	database := os.Getenv("CLICKHOUSE_DB")
	user := os.Getenv("CLICKHOUSE_USER")
	password := os.Getenv("CLICKHOUSE_PASSWORD")

	ch, _ = sql.Open(
		"clickhouse",
		fmt.Sprintf("tcp://%s:%s?username=%s&password=%s&database=%s", host, port, user, password, database),
	)
}
func TestMain(m *testing.M) {
	setUp()
	defer rClient.Close()
	defer ch.Close()

	os.Exit(m.Run())
}

func TestRedisConnection(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	pong, err := rClient.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	t.Logf("Подключение к редис успешно: %s", pong)
}

func TestClickHouseConnection(t *testing.T) {
	err := ch.Ping()
	if err != nil {
		t.Fatalf("Не удалось выполнить ping ClickHouse: %v", err)
	}

	t.Log("Подключение к ClickHouse успешно")
}
