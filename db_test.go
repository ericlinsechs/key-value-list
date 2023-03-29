package main

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConnectToDB(t *testing.T) {
	// Define the test PostgreSQL server connection parameters
	host := "localhost"
	port := "5432"
	user := "myuser"
	password := "mysecretpassword"
	dbname := "my_database"

	// Define the connection string for the PostgreSQL server
	connStr := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"

	// Connect to the PostgreSQL server
	expectedDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %s", err)
	}

	// Connect to the test PostgreSQL server and get the actual GORM DB object
	actualDB, err := connectToDB(host, port, user, password, dbname)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Check if the actual DB object matches the expected DB object
	if actualDB.Name() != expectedDB.Name() {
		t.Errorf("DB names do not match: expected %s, but got %s", expectedDB.Name(), actualDB.Name())
	}

	// Close the database connections
	actualDBSQL, _ := actualDB.DB()
	actualDBSQL.Close()

	expectedDBSQL, _ := expectedDB.DB()
	expectedDBSQL.Close()
}

func TestCreateDatabase(t *testing.T) {
	// Define test data
	host := "localhost"
	port := "5432"
	user := "myuser"
	password := "mysecretpassword"
	// dbname := "my_database"
	dbname := "testdb"

	// Connect to mock PostgreSQL server
	mockDB, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", host, port, user, password))
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer mockDB.Close()

	// Call createDatabase function
	err = createDatabase(host, port, user, password, dbname)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that database was created
	rows, err := mockDB.Query("SELECT 1 FROM pg_database WHERE datname=$1", dbname)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Errorf("Database %s was not created", dbname)
	}

	// Call createDatabase function again with same dbname
	err = createDatabase(host, port, user, password, dbname)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that no new databases were created
	rows, err = mockDB.Query("SELECT COUNT(*) FROM pg_database WHERE datname LIKE 'testdb%'")
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			t.Fatalf("Error scanning rows: %v", err)
		}
	}
	if count != 1 {
		t.Errorf("Unexpected number of databases created: %d", count)
	}
}
