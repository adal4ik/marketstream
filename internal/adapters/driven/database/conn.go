package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"marketstream/internal/config"
)

func ConnectDB(cfg config.DatabaseConfig) *sql.DB {
	time.Sleep(5 * time.Second)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed, dont ponged: %v", err)
	}

	log.Println("Successfully connected to the database!")
	return db
}
