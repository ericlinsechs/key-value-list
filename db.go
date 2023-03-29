package main

import (
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

	sqlDB, _ := db.DB()

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Check if the connection is successful
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// func connectToDB(driver, host, port, user, password, dbname string) (*gorm.DB, error) {
// 	var dialect gorm.Dialector

// 	switch driver {
// 	case "postgres":
// 		// Define the connection string for the PostgreSQL server
// 		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

// 		// Set the PostgreSQL driver
// 		dialect = postgres.Open(connStr)
// 	case "sqlite":
// 		// Define the connection string for the SQLite database
// 		// connStr := fmt.Sprintf("%s.db", dbname)

// 		// Set the SQLite driver
// 		dialect = sqlite.Open(":memory:")
// 	default:
// 		// Return an error if the specified driver is not supported
// 		return nil, errors.New("unsupported database driver")
// 	}

// 	// Connect to the database
// 	db, err := gorm.Open(dialect, &gorm.Config{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	sqlDB, _ := db.DB()

// 	sqlDB.SetMaxIdleConns(5)
// 	sqlDB.SetMaxOpenConns(20)
// 	sqlDB.SetConnMaxLifetime(5 * time.Minute)

// 	// Check if the connection is successful
// 	err = sqlDB.Ping()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return db, nil
// }

// CreateDatabase creates a new database with the given name if it does not already exist.
func createDatabase(host string, port string, user string, password string, dbname string) error {
	// Define the connection string for the PostgreSQL server
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", host, port, user, password)

	// Connect to the PostgreSQL server
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, _ := db.DB()

	defer sqlDB.Close()

	// Check if database exists
	var exists bool
	err = sqlDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dbname).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Database '%s' already exists\n", dbname)
		return nil
	}

	// Database does not exist, create it
	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		return err
	}

	log.Printf("Database '%s' created successfully\n", dbname)

	return nil
}
