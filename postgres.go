package main

import (
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connectToDB(host string, port string, user string, password string, dbname string) (*gorm.DB, error) {
	// Define the connection string for the PostgreSQL server
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Connect to the PostgreSQL server
	// db, err := gorm.Open("postgres", connStr)
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// createDatabaseIfNotExists creates a new database with the given name if it does not already exist.
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

// func getPageWithArticles(pageID uint) (Page, error) {
// 	var page Page
// 	if err := db.Preload("Articles").First(&page, pageID).Error; err != nil {
// 		return Page{}, err
// 	}
// 	return page, nil
// }

func getPageByID(db *gorm.DB, pageID uint, page *Page) error {
	if err := db.Table("pages").First(page, pageID).Error; err != nil {
		return err
	}
	return nil
}

// GetLastPage gets the last Page in the database.
func getLastPage(db *gorm.DB, lastPage *Page) error {
	return db.Table("pages").Last(lastPage).Error
}

func getArticlesByPageID(db *gorm.DB, pageID uint, articles *[]Article) error {
	if err := db.Table("articles").Where("page_id = ?", pageID).Find(articles).Error; err != nil {
		return err
	}
	return nil
}

// PreloadArticles preloads the Articles associated with the Page in the database.
func preloadArticles(db *gorm.DB, page *Page) error {
	return db.Preload("Articles").First(page, page.ID).Error
}

// SavePage saves the given Page to the database.
func savePage(db *gorm.DB, page *Page) error {
	return db.Save(page).Error
}

// CreatePage creates a new Page in the database.
func createPage(db *gorm.DB, page *Page) error {
	return db.Table("pages").Create(page).Error
}

// UpdateLastPageNextPageID updates the NextPageID of the last Page in the database.
func updateLastPageNextPageID(db *gorm.DB, lastPage *Page, newPageID uint) error {
	return db.Model(lastPage).UpdateColumn("next_page_id", newPageID).Error
}

// SaveArticle saves the given Article to the database.
func saveArticle(db *gorm.DB, article *Article) error {
	return db.Save(article).Error
}
