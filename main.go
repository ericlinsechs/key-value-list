package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// Define the Article model with a primary key
type Article struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Title     string
	Author    string
	Content   string
	PageID    uint // foreign key to Page.ID
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

const (
	host     = "localhost"
	port     = "5432"
	user     = "postgres"
	password = "mysecretpassword"
	dbname   = "my_database"
)

const (
	FirstListKey = 1
	FirstPageKey = 1
)

const (
	ArticleSample = 100
)

var db *gorm.DB
var conn *sql.DB

func init() {
	var err error
	// Connect to the PostgreSQL server
	db, err = connectToDB(host, port, user, password, "postgres")
	if err != nil {
		panic(err)
	}

	// Create a new database if it does not exist
	conn, _ = db.DB()
	if err = createDatabaseIfNotExists(conn, dbname); err != nil {
		log.Fatal(err)
	}

	// Use the new database
	db, err = connectToDB(host, port, user, password, dbname)
	if err != nil {
		panic(err)
	}

	// Auto-migrate the schema to create the tables and relationships
	if err = db.AutoMigrate(&Article{}, &Page{}); err != nil {
		// Handle error here
		log.Fatalf("Error during migration: %v", err)
	}

}

func main() {
	defer conn.Close()

	// createSampleArticle()

	r := mux.NewRouter()

	// list
	// r.HandleFunc("/list/{list_id}", getHead).Methods("GET")

	// page
	r.HandleFunc("/page/get/{page_id}", getPage).Methods("GET")
	r.HandleFunc("/page/set/{page_id}", set).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))
}

func getPage(w http.ResponseWriter, r *http.Request) {
	var err error
	var id string

	vars := mux.Vars(r)
	id = vars["page_id"]

	pageID64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		// handle error
	}
	pageID := uint(pageID64)

	var page Page
	// where "id" is the ID of the page you want to retrieve
	err = getPageByID(db, pageID, &page)
	if err != nil {
		log.Fatal(err)
	}

	// Query the database to get the articles associated with the specified page ID
	var articles []Article
	err = getArticlesByPageID(db, pageID, &articles)
	if err != nil {
		log.Fatal(err)
	}

	var articleData []map[string]string
	for _, article := range articles {
		articleData = append(articleData, map[string]string{
			"title":   article.Title,
			"author":  article.Author,
			"content": article.Content,
		})
	}
	res := map[string]interface{}{
		"articles":     articleData,
		"next_page_id": page.NextPageID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func set(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the article data
	var article Article
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// func set(article Article) {
	addArticleToPage(article)
}

func createNewPage(newPageID uint, ListID uint) (page Page) {
	var err error
	// Create a new page with the given pageID and ListID
	page = Page{
		ID:     newPageID,
		ListID: ListID,
	}
	if newPageID != FirstPageKey {
		var lastPage Page
		err = getLastPage(db, &page)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("last page: %+v", lastPage)
		err = updateLastPageNextPageID(db, &lastPage, newPageID)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = createPage(db, &page)
	if err != nil {
		log.Fatal(err)
	}
	return page
}

func addArticleToPage(newArticle Article) error {
	var page Page
	var err error

	// find the page you want to add the article to
	err = getLastPage(db, &page)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Pages table is empty
			fmt.Println("create the first page")
			page = createNewPage(FirstPageKey, FirstListKey)
		} else {
			log.Fatal(err)
		}
	}

	err = preloadArticles(db, &page)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("article len: %v\n", len(page.Articles)+1)

	if len(page.Articles) >= 5 {
		page = createNewPage(page.ID+1, page.ListID)
		log.Printf("createNewPage id: %v\n", page.ID)
	}

	// append the new article to the Articles slice of the page
	page.Articles = append(page.Articles, newArticle)

	// Save the updated page back to the database
	err = savePage(db, &page)
	if err != nil {
		log.Fatal(err)
	}

	err = preloadArticles(db, &page)
	if err != nil {
		log.Fatal(err)
	}

	for _, article := range page.Articles {
		fmt.Printf("Article ID: %d, Title: %s, Author: %s, Content: %s\n", article.ID, article.Title, article.Author, article.Content)
	}
	return nil
}

// func createSampleArticle() {

// 	// Define a slice of Article structs to hold the records to be inserted
// 	for i := 1; i <= ArticleSample; i++ {
// 		article := Article{
// 			Title:   fmt.Sprintf("Article %d", i),
// 			Author:  fmt.Sprintf("Author %d", i),
// 			Content: fmt.Sprintf("Content %d", i),
// 		}
// 		set(article)
// 	}
// }
