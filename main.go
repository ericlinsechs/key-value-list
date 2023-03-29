package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

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
	ArticleSample            = 21
	NumberOfArticleInOnePage = 5
)

var db *gorm.DB
var sqlDB *sql.DB

func initDB() {
	var err error

	// Create the database if it doesn't exist
	err = createDatabase(host, port, user, password, dbname)
	if err != nil {
		fmt.Println("Error creating database:", err)
		return
	}

	// Connect to the PostgreSQL server
	db, err = connectToDB(host, port, user, password, dbname)
	if err != nil {
		panic(err)
	}

	// Auto-migrate the schema to create the tables and relationships
	if err = db.AutoMigrate(&Page{}, &Article{}, &List{}); err != nil {
		// Handle error here
		log.Fatalf("Error during migration: %v", err)
	}

	createListIfNotExists()

	// Create same sample articles if there is no data in articles table
	createSampleArticle()
}

func main() {
	initDB()

	sqlDB, _ = db.DB()
	defer sqlDB.Close()

	r := mux.NewRouter()

	// list
	r.HandleFunc("/list/get", handleGetHead).Methods("GET")

	// page
	r.HandleFunc("/page/get", handleGetPage).Methods("GET")
	r.HandleFunc("/page/set", handleSet).Methods("POST")
	r.HandleFunc("/page/update", handleUpdate).Methods("POST")
	r.HandleFunc("/page/delete", handleDeletePage).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", r))
}

func handleGetHead(w http.ResponseWriter, r *http.Request) {
	if err := getHead(w, r); err != nil {
		log.Printf("Error in getHead: %v\n", err)
		if strings.Contains(err.Error(), "missing") || strings.Contains(err.Error(), "valid integer") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func handleGetPage(w http.ResponseWriter, r *http.Request) {
	if err := getPage(w, r); err != nil {
		log.Printf("Error in getPage: %v\n", err)
		if strings.Contains(err.Error(), "missing") || strings.Contains(err.Error(), "valid integer") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if err := set(w, r); err != nil {
		log.Printf("Error in set: %v\n", err)
		if strings.Contains(err.Error(), "invalid request body") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if err := update(w, r); err != nil {
		log.Printf("Error in update: %v\n", err)
		if strings.Contains(err.Error(), "page_id parameter is missing") ||
			strings.Contains(err.Error(), "page_id parameter is not a valid integer") ||
			strings.Contains(err.Error(), "Invalid request body") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "Page not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func handleDeletePage(w http.ResponseWriter, r *http.Request) {
	if err := deletePage(w, r); err != nil {
		log.Printf("Error in deletePage: %v\n", err)
		if strings.Contains(err.Error(), "list_id parameter is missing") ||
			strings.Contains(err.Error(), "list_id parameter is not a valid integer") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
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
		err = getLastPage(db, &lastPage)
		if err != nil {
			log.Fatal(err)
		}
		// log.Printf("last page: %+v", lastPage)
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

	if len(page.Articles) >= NumberOfArticleInOnePage {
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

	// err = preloadArticles(db, &page)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Printf("Add Article to page id: %v\n", page.ID)
	// for _, article := range page.Articles {
	// 	log.Printf("Article ID: %d, Title: %s, Author: %s, Content: %s\n", article.ID, article.Title, article.Author, article.Content)
	// }
	return nil
}

func createSampleArticle() {
	// Check if any records exist in the articles table
	var count int64
	if err := db.Model(&Article{}).Count(&count).Error; err != nil {
		log.Fatalf("Error checking if records exist in articles table: %v", err)
	}

	// If there are no records, create the sample articles
	if count == 0 {
		// Define a slice of Article structs to hold the records to be inserted
		for i := 1; i <= ArticleSample; i++ {
			article := Article{
				Title:   fmt.Sprintf("Article %d", i),
				Author:  fmt.Sprintf("Author %d", i),
				Content: fmt.Sprintf("Content %d", i),
			}
			if err := addArticleToPage(article); err != nil {
				log.Fatalf("Error adding article to page: %v", err)
			}
		}
	}
}

func createListIfNotExists() error {
	var list List
	err := getListByID(db, 1, &list)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// create a new list if none exists
			list = List{
				ID:         FirstListKey,
				NextPageID: FirstPageKey,
			}
			err = createList(db, &list)
			if err != nil {
				return err
			}
			log.Printf("Created new list: %+v\n", list)
			return nil
		}
		// return any other errors
		return err
	}
	log.Printf("Found existing list: %+v\n", list)
	return nil
}
