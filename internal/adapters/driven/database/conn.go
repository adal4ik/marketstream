package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"marketstream/internal/config"
)

func ConnectDB(cfg config.DatabaseConfig) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	var db *sql.DB
	var err error

	maxRetries := 10
	retryDelay := 2 * time.Second

	for i := 1; i <= maxRetries; i++ {
		db, err = sql.Open("pgx", psqlInfo)
		if err != nil {
			log.Printf("[Attempt %d/%d] Failed to open DB: %v", i, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

		err = db.Ping()
		if err == nil {
			log.Println("Successfully connected to the database!")
			return db
		}
		
		log.Printf("[Attempt %d/%d] Database ping failed: %v", i, maxRetries, err)
		time.Sleep(retryDelay)
	}

	log.Fatalf("Database unreachable after %d attempts: %v", maxRetries, err)
	return nil
}
