package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	host        = os.Getenv("DB_HOST")
	user        = os.Getenv("DB_USER")
	password    = os.Getenv("DB_PASSWORD")
	name        = os.Getenv("DB_NAME")
	port, exist = os.LookupEnv("DB_PORT")
)

func ConnectDB() *sql.DB {
	time.Sleep(5 * time.Second)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, name)

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
