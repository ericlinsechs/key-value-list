package main

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetListByID(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Automatically create the "lists" table in the database
	err = db.AutoMigrate(&List{})
	if err != nil {
		t.Fatalf("Failed to migrate the database: %v", err)
	}

	// Create a new list and save it to the database
	list := &List{ID: 1}
	err = createList(db, list)
	if err != nil {
		t.Fatalf("Failed to create a new list: %v", err)
	}

	// Retrieve the list by ID from the database
	var retrievedList List
	err = getListByID(db, 1, &retrievedList)
	if err != nil {
		t.Fatalf("Failed to retrieve the list by ID: %v", err)
	}

	// Compare the original list and the retrieved list
	if retrievedList.ID != list.ID {
		t.Errorf("Retrieved list ID (%d) does not match the original list ID (%d)", retrievedList.ID, list.ID)
	}
}

func TestCreateList(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Automatically create the "lists" table in the database
	err = db.AutoMigrate(&List{})
	if err != nil {
		t.Fatalf("Failed to migrate the database: %v", err)
	}

	// Create a new list and save it to the database
	list := &List{ID: 1}
	err = createList(db, list)
	if err != nil {
		t.Fatalf("Failed to create a new list: %v", err)
	}

	// Retrieve the list by ID from the database
	var retrievedList List
	err = getListByID(db, 1, &retrievedList)
	if err != nil {
		t.Fatalf("Failed to retrieve the list by ID: %v", err)
	}

	// Compare the original list and the retrieved list
	if retrievedList.ID != list.ID {
		t.Errorf("Retrieved list ID (%d) does not match the original list ID (%d)", retrievedList.ID, list.ID)
	}
}
