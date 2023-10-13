package main

import (
	"database/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
	"os"
)

var i18nDB *sql.DB

func init() {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{os.Getenv("CLICKHOUSE_DB_URL")},
		Auth: clickhouse.Auth{
			Database: "i18n",
			Username: os.Getenv("CLICKHOUSE_DB_USER"),
			Password: os.Getenv("CLICKHOUSE_DB_PASSWORD"),
		},
		Protocol: clickhouse.HTTP,
	})

	i18nDB = conn
}
