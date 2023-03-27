package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Define the Article model with a primary key
type Article struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Title     string     `json:"title"`
	Author    string     `json:"author"`
	Content   string     `json:"content"`
	PageID    uint       // foreign key to Page.ID
}

// Define the Page model with a foreign key to the Article model
type Page struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `sql:"index"`
	ListID     uint
	Articles   []Article `gorm:"ForeignKey:PageID"`
	NextPageID uint
}

type List struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `sql:"index"`
	NextPageID uint
}

// ConnectToDB connects to the PostgreSQL server and returns a GORM DB object.
func connectToDB(host string, port string, user string, password string, dbname string) (*gorm.DB, error) {
	// Define the connection string for the PostgreSQL server
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Connect to the PostgreSQL server
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	conn, _ = db.DB()

	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(20)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Check if the connection is successful
	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CreateDatabaseIfNotExists creates a new database with the given name if it does not already exist.
func createDatabaseIfNotExists(conn *sql.DB, dbName string) error {
	// Check if the database already exists
	rows, err := conn.Query(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", dbName))
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		// Database already exists, no need to create it
		return nil
	}

	// Database does not exist, create it
	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return err
	}
	log.Println("Database created successfully")

	return nil
}
