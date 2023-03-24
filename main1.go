package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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

var db *gorm.DB

func init() {
	var err error
	// Connect to the PostgreSQL server
	db, err = connectToDB(host, port, user, password, "postgres")
	if err != nil {
		panic(err)
	}

	// Create a new database if it does not exist
	if err = createDatabaseIfNotExists(db.DB(), dbname); err != nil {
		log.Fatal(err)
	}

	// Use the new database
	db, err = connectToDB(host, port, user, password, dbname)
	if err != nil {
		panic(err)
	}

	// Auto-migrate the schema to create the tables and relationships
	if err = db.AutoMigrate(&Article{}, &Page{}).Error; err != nil {
		// Handle error here
		log.Fatalf("Error during migration: %v", err)
	}

}

func main() {
	defer db.Close()

	// var pages []Page
	// var articles []Article
	// db.Find(&articles)
	// fmt.Printf("%+v\n", articles)
	// for _, article := range articles {
	// 	fmt.Printf("%+v\n", article)
	// }
	set()

	// r := mux.NewRouter()

	// list
	// r.HandleFunc("/list/{list_id}", getHead).Methods("GET")

	// page
	// r.HandleFunc("/page/getall", getAllPage).Methods("GET")
	// r.HandleFunc("/page/get/{page_id}", getPage).Methods("GET")
	// r.HandleFunc("/page/set/{page_id}", set).Methods("POST")

	// // article
	// r.HandleFunc("/article/getall", getAllArticle).Methods("GET")
	// r.HandleFunc("/article/create", createArticle).Methods("POST")

	// log.Fatal(http.ListenAndServe(":8000", r))
}

func connectToDB(host string, port string, user string, password string, dbname string) (*gorm.DB, error) {
	// Define the connection string for the PostgreSQL server
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Connect to the PostgreSQL server
	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// createDatabaseIfNotExists creates a new database with the given name if it does not already exist.
func createDatabaseIfNotExists(db *sql.DB, dbName string) error {
	// Check if the database already exists
	rows, err := db.Query(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", dbName))
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		// Database already exists, no need to create it
		return nil
	}

	// Database does not exist, create it
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return err
	}
	log.Println("Database created successfully")

	return nil
}

func getAllPage(w http.ResponseWriter, r *http.Request) {
	var pages []Page

	db.Table("pages").Find(&pages)
	// fmt.Printf("%+v\n", pages)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(pages)
	if err != nil {
		log.Fatal(err)
	}
}

func getPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["page_id"]

	var page Page
	db.Table("pages").First(&page, pageID) // where "id" is the ID of the page you want to retrieve

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page)
}

// func set(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	pageID, err := primitive.ObjectIDFromHex(vars["page_id"])
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	log.Printf("input pageID: %q\n", pageID)

// 	// Get the articles from the database
// 	var articles []Article
// 	db.Order(gorm.Expr("RANDOM()")).Limit(5).Find(&articles)

func set() {
	var FirstPageKey uint = 1
	var FirstListKey uint = 1
	// Define a slice of Article structs to hold the records to be inserted
	// i := 0
	// i++
	// newArticle := Article{
	// 	Title:   fmt.Sprintf("Article %d", i),
	// 	Author:  fmt.Sprintf("Author %d", i),
	// 	Content: fmt.Sprintf("Content %d", i),
	// }

	// Check if the Pages table is empty
	var page Page
	if err := db.Table("pages").First(&page).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Pages table is empty
			fmt.Println("create the first page")

			// Create a new page with the given pageID and lastPage.ListID
			page = Page{
				ID:     FirstPageKey,
				ListID: FirstListKey,
			}

			db.Table("pages").Create(&page)
		} else {
			log.Fatal(err)
		}
	} else {
		// Pages table is not empty
		// Get the last page in the database
		if err := db.Table("pages").First(&page).Error; err != nil {
			log.Fatal(err)
		}
	}

	// log.Printf("%+v", page)
	// addArticleToPage(1, newArticle)

	// create a new Article
	newArticle := Article{
		Title:   "New Article Title",
		Author:  "New Article Author",
		Content: "New Article Content",
		PageID:  1, // set the PageID to the ID of the page you want to add it to
	}

	// find the page you want to add the article to
	var p Page
	if err := db.First(&p, 1).Error; err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", p)

	// append the new article to the Articles slice of the page
	p.Articles = append(p.Articles, newArticle)

	// // save the updated page to the database
	// if err := db.Save(&p).Error; err != nil {
	// 	log.Fatal(err)
	// }
	// Update existing page record
	if err := db.Model(&Page{}).Where("id = ?", p.ID).Updates(p).Error; err != nil {
		// Handle the error
		log.Fatalf("Failed to update article: %v", err)
	}

	if err := db.Preload("Articles").First(&page, p.ID).Error; err != nil {
		// Handle error
		log.Fatalf("Failed query page: %v", err)
	}

	for _, article := range page.Articles {
		fmt.Printf("Article ID: %d, Title: %s, Author: %s, Content: %s\n", article.ID, article.Title, article.Author, article.Content)
	}
}

func addArticleToPage(pageID uint, article Article) error {
	// Query the page from the database
	var page Page
	if err := db.First(&page, pageID).Error; err != nil {
		return fmt.Errorf("failed to query page: %v", err)
	}

	// Add the article to the page's Articles slice
	page.Articles = append(page.Articles, article)

	// Save the page back to the database
	if err := db.Save(&page).Error; err != nil {
		return fmt.Errorf("failed to save page: %v", err)
	}

	return nil
}

func getAllArticle(w http.ResponseWriter, r *http.Request) {
	var articles []Article

	db.Table("articles").Find(&articles)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(articles)
	if err != nil {
		log.Fatal(err)
	}
}

func createArticle(w http.ResponseWriter, r *http.Request) {
	var articlesToInsert []Article

	// Define a slice of Article structs to hold the records to be inserted
	for i := 1; i <= 100; i++ {
		article := Article{
			Title:   fmt.Sprintf("Article %d", i),
			Author:  fmt.Sprintf("Author %d", i),
			Content: fmt.Sprintf("Content %d", i),
		}
		articlesToInsert = append(articlesToInsert, article)
	}

	// Use the Create method of the GORM instance to insert the records into the database
	if err := db.Table("articles").Create(&articlesToInsert).Error; err != nil {
		log.Fatal(err)
	}
}
