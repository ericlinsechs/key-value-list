package main

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSaveArticle(t *testing.T) {
	// create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, _ := db.DB()

	defer sqlDB.Close()

	// create necessary tables for testing
	err = db.AutoMigrate(&Article{})
	if err != nil {
		t.Fatal(err)
	}

	// create a test article
	article := &Article{
		Title:   "Test Article",
		Author:  "Test Author",
		Content: "Test Content",
	}

	// save the article to the database
	if err := saveArticle(db, article); err != nil {
		t.Fatal(err)
	}

	// retrieve the article from the database to ensure it was saved
	var retrievedArticle Article
	if err := db.First(&retrievedArticle, article.ID).Error; err != nil {
		t.Fatal(err)
	}

	// ensure the retrieved article matches the original article
	if retrievedArticle.Title != article.Title {
		t.Errorf("expected title %q, but got %q", article.Title, retrievedArticle.Title)
	}
	if retrievedArticle.Author != article.Author {
		t.Errorf("expected author %q, but got %q", article.Author, retrievedArticle.Author)
	}
	if retrievedArticle.Content != article.Content {
		t.Errorf("expected content %q, but got %q", article.Content, retrievedArticle.Content)
	}
}

func TestDeleteArticlesByPageID(t *testing.T) {
	// create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, _ := db.DB()

	defer sqlDB.Close()

	// create necessary tables for testing
	err = db.AutoMigrate(&Article{}, &Page{})
	if err != nil {
		t.Fatal(err)
	}

	// create a test page
	page := &Page{}
	if err := db.Create(page).Error; err != nil {
		t.Fatal(err)
	}

	// create some test articles associated with the page
	for i := 0; i < 3; i++ {
		article := &Article{
			Title:   "Test Article",
			Author:  "Test Author",
			Content: "Test Content",
			PageID:  page.ID,
		}
		if err := db.Create(article).Error; err != nil {
			t.Fatal(err)
		}
	}

	// delete the articles associated with the page
	if err := deleteArticlesByPageID(db, page.ID); err != nil {
		t.Fatal(err)
	}

	// ensure the articles were deleted
	var count int64
	if err := db.Model(&Article{}).Where("page_id = ?", page.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 articles associated with page ID %d, but got %d", page.ID, count)
	}
}
