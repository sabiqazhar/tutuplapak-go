package provider

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(dataSourceName string) *sql.DB {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return db
}