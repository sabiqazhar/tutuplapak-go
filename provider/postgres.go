package provider

import (
	"database/sql"
	"fmt"
	"log"

	"tutuplapak-go/config"

	_ "github.com/lib/pq"
)

func InitDB(cfg config.DBConfig) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name)
	fmt.Println(dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return db
}
