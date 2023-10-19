package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
)

var i18nDB *pgxpool.Pool

func init() {
	var err error
	i18nDB, err = pgxpool.New(context.Background(), os.Getenv("POSTGRES_DB_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	createTables()
}

func createTables() {
	q, err := i18nDB.Query(context.Background(), `
CREATE TABLE IF NOT EXISTS public
(
    "group" varchar,
    "key"   varchar,
    "en_us" varchar,
    "ru_ru" varchar,
    
    UNIQUE ("key", "group")
)
`)
	if err != nil {
		return
	}

	q.Close()
}
