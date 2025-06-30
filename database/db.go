package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	log.Println("Successfully connected to the database!")
	createTables()
}

func createTables() {
	createJobPostingsTableSQL := `
	CREATE TABLE IF NOT EXISTS job_postings (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL
	);`

	createApplicantsTableSQL := `
	CREATE TABLE IF NOT EXISTS applicants (
		id SERIAL PRIMARY KEY,
		job_id TEXT NOT NULL REFERENCES job_postings(id),
		name TEXT NOT NULL,
		grade NUMERIC(5,2) NOT NULL,
		skills TEXT[],
		description TEXT NOT NULL
	);`

	_, err := DB.Exec(createJobPostingsTableSQL)
	if err != nil {
		log.Fatal("Error creating job_postings table:", err)
	}

	_, err = DB.Exec(createApplicantsTableSQL)
	if err != nil {
		log.Fatal("Error creating applicants table:", err)
	}

	log.Println("Tables created successfully or already exist.")
}