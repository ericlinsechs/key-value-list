package main

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetPageByID(t *testing.T) {
	// Initialize a new in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// Migrate the database schema
	db.AutoMigrate(&List{}, &Page{}, &Article{})

	// Create a new Page and save it to the database
	page := &Page{ListID: 1, NextPageID: 0}
	if err := createPage(db, page); err != nil {
		t.Fatal(err)
	}

	// Get the Page by ID
	var pageResult Page
	if err := getPageByID(db, page.ID, &pageResult); err != nil {
		t.Fatal(err)
	}
	// Check if the retrieved Page is the same as the original Page
	if pageResult.ID != page.ID || pageResult.ListID != page.ListID || pageResult.NextPageID != page.NextPageID {
		t.Errorf("Retrieved page does not match original page")
	}
}

func TestGetLastPage(t *testing.T) {
	// Initialize a new in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// Migrate the database schema
	db.AutoMigrate(&List{}, &Page{}, &Article{})

	// Create two new Pages and save them to the database
	page1 := &Page{ListID: 1, NextPageID: 2}
	page2 := &Page{ListID: 1, NextPageID: 0}
	if err := createPage(db, page1); err != nil {
		t.Fatal(err)
	}
	if err := createPage(db, page2); err != nil {
		t.Fatal(err)
	}

	// Get the last Page in the database
	var lastPage Page
	if err := getLastPage(db, &lastPage); err != nil {
		t.Fatal(err)
	}
	// Check if the retrieved Page is the same as the last created Page
	if lastPage.ID != page2.ID || lastPage.ListID != page2.ListID || lastPage.NextPageID != page2.NextPageID {
		t.Errorf("Retrieved last page does not match last created page")
	}
}

func TestGetArticlesByPageID(t *testing.T) {
	// Initialize a new in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// Migrate the database schema
	err = db.AutoMigrate(&Article{})
	if err != nil {
		t.Fatalf("Failed to migrate the database schema: %v", err)
	}

	// Create some test data
	pageID := uint(1)
	articles := []Article{
		{Title: "Article 1", PageID: pageID},
		{Title: "Article 2", PageID: pageID},
		{Title: "Article 3", PageID: pageID},
	}
	err = db.Create(&articles).Error
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	// Call the function being tested
	var result []Article
	err = getArticlesByPageID(db, pageID, &result)
	if err != nil {
		t.Fatalf("Function returned an error: %v", err)
	}

	// Check the result
	if len(result) != len(articles) {
		t.Fatalf("Unexpected result length. Expected %d, got %d", len(articles), len(result))
	}
	for i, article := range articles {
		if article.Title != result[i].Title {
			t.Fatalf("Unexpected article title. Expected %q, got %q", article.Title, result[i].Title)
		}
		if article.PageID != result[i].PageID {
			t.Fatalf("Unexpected article page ID. Expected %d, got %d", article.PageID, result[i].PageID)
		}
	}
}
func TestPreloadArticles(t *testing.T) {
	// Initialize a new in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// Migrate the schema
	db.AutoMigrate(&Page{}, &Article{})

	// Create a test Page with associated Articles
	page := &Page{ListID: 1, NextPageID: 0}
	db.Create(page)
	article1 := &Article{Title: "Test Article 1", Author: "John Doe", Content: "Test content 1", PageID: page.ID}
	db.Create(article1)
	article2 := &Article{Title: "Test Article 2", Author: "Jane Smith", Content: "Test content 2", PageID: page.ID}
	db.Create(article2)

	// Preload the Articles for the test Page
	err = preloadArticles(db, page)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that the Articles are correctly preloaded
	if len(page.Articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(page.Articles))
	}
	if page.Articles[0].Title != article1.Title {
		t.Errorf("expected article 1 title to be %q, got %q", article1.Title, page.Articles[0].Title)
	}
	if page.Articles[1].Title != article2.Title {
		t.Errorf("expected article 2 title to be %q, got %q", article2.Title, page.Articles[1].Title)
	}
}

func TestCreatePage(t *testing.T) {
	// Initialize a new in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	// Automatically create the "pages" table
	err = db.AutoMigrate(&Page{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	// Create a new page
	page := &Page{
		ListID:     99,
		NextPageID: 100,
	}
	err = createPage(db, page)
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}

	// Retrieve the page from the database
	var retrievedPage Page
	err = db.Table("pages").First(&retrievedPage, page.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve page: %v", err)
	}

	// Compare the retrieved page with the original page
	if page.ListID != retrievedPage.ListID || page.NextPageID != retrievedPage.NextPageID {
		t.Errorf("Retrieved page does not match original page:\nExpected: %v\nActual: %v", *page, retrievedPage)
	}
}

func TestUpdateLastPageNextPageID(t *testing.T) {
	// Initialize a new in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	// Automatically create the "pages" table
	err = db.AutoMigrate(&Page{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	// Create two new pages
	firstPage := &Page{
		ListID:     99,
		NextPageID: 0,
	}
	err = createPage(db, firstPage)
	if err != nil {
		t.Fatalf("Failed to create first page: %v", err)
	}
	secondPage := &Page{
		ListID:     100,
		NextPageID: 0,
	}
	err = createPage(db, secondPage)
	if err != nil {
		t.Fatalf("Failed to create second page: %v", err)
	}

	// Update the NextPageID of the first page to point to the second page
	err = updateLastPageNextPageID(db, firstPage, secondPage.ID)
	if err != nil {
		t.Fatalf("Failed to update NextPageID: %v", err)
	}

	// Retrieve the first page from the database
	var retrievedFirstPage Page
	err = db.Table("pages").First(&retrievedFirstPage, firstPage.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve first page: %v", err)
	}

	// Compare the NextPageID of the first page with the ID of the second page
	if retrievedFirstPage.NextPageID != secondPage.ID {
		t.Errorf("NextPageID does not match expected value:\nExpected: %v\nActual: %v", secondPage.ID, retrievedFirstPage.NextPageID)
	}
}

func TestSavePage(t *testing.T) {
	// Initialize a new in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Automatically create the "pages" table
	err = db.AutoMigrate(&Page{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	// Create a new test page
	page := &Page{
		ListID:     100,
		NextPageID: 101,
	}

	// Save the test page to the database
	err = savePage(db, page)
	if err != nil {
		t.Fatalf("Failed to save page: %v", err)
	}

	// Retrieve the saved page from the database
	retrievedPage := &Page{}
	err = db.Table("pages").First(retrievedPage, page.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve page: %v", err)
	}

	// Compare the retrieved page with the original page
	if retrievedPage.ListID != page.ListID || retrievedPage.NextPageID != page.NextPageID {
		t.Errorf("Retrieved page does not match original page:\nExpected: %v\nActual: %v", page, retrievedPage)
	}
}

func TestDeletePagesByListID(t *testing.T) {
	// Initialize a new in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Migrate the database schema
	err = db.AutoMigrate(&List{}, &Page{}, &Article{})
	if err != nil {
		t.Fatalf("Failed to migrate the database schema: %v", err)
	}

	// Create a test list and pages
	testList := List{ID: 1, NextPageID: 2}
	testPages := []Page{
		{ID: 1, ListID: 1},
		{ID: 2, ListID: 1},
	}
	if err := db.Table("lists").Create(&testList).Error; err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}
	for i := range testPages {
		if err := db.Table("pages").Create(&testPages[i]).Error; err != nil {
			t.Fatalf("Failed to create test page: %v", err)
		}
	}

	// Delete the test pages
	if err := deletePagesByListID(db, 1); err != nil {
		t.Fatalf("Failed to delete pages: %v", err)
	}

	// Verify that the pages have been deleted
	var count int64
	if err := db.Table("pages").Where("list_id = ?", 1).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count pages: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 pages after deletion, but got %d", count)
	}
}
