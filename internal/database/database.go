package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func Connect(dsn string) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	// test connection
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	log.Println("🐘 Connected to PostgreSQL successfully!")

	return db
}
