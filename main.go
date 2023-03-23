package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Define the Article model with a primary key
type Article struct {
	gorm.Model
	Title   string
	Author  string
	Content string
}

// Define the Page model with a foreign key to the Article model
type Page struct {
	gorm.Model
	ListID     uint
	Articles   []*Article `gorm:"ForeignKey:PageID"`
	NextPageID uint
}

func main() {
	// Define the connection string for the PostgreSQL database
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=postgres password=mysecretpassword dbname=my_database sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Auto-migrate the schema to create the tables and relationships
	db.AutoMigrate(&Article{}, &Page{})

	// Create a new article
	article := &Article{
		Title:   "New Article",
		Author:  "John Doe",
		Content: "This is a new article",
	}
	db.Create(article)

	// Create a new page and associate it with the article
	page := &Page{
		ListID:     1, // The ID of the list this page belongs to
		NextPageID: 2, // The ID of the next page in the list
	}
	db.Create(page)

	// Associate the article with the page
	db.Model(page).Association("Articles").Append(article)

	// Retrieve the page and its associated articles
	var pages []Page
	db.Preload("Articles").Find(&pages)
	fmt.Printf("%+v\n", pages)
}
