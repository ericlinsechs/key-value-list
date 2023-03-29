package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	ArticleSample = 21
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
	r.HandleFunc("/page/set", set).Methods("POST")
	r.HandleFunc("/page/update/{page_id}", update).Methods("POST")
	r.HandleFunc("/page/delete/{list_id}", deletePage).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", r))
}

func handleGetHead(w http.ResponseWriter, r *http.Request) {
	if err := getHead(w, r); err != nil {
		log.Printf("Error in getHead: %v\n", err)
		if strings.Contains(err.Error(), "missing") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func getHead(w http.ResponseWriter, r *http.Request) error {
	// Extract the value of the "list_id" query parameter
	idStr := r.URL.Query().Get("list_id")
	if idStr == "" {
		return fmt.Errorf("list_id parameter is missing")
	}
	// Validate the list ID parameter
	listID, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("list_id parameter is not a valid integer")
	}

	if db == nil { // check if db is nil
		return fmt.Errorf("database connection is nil")
	}

	// Fetch the list from the database
	var list List
	if err := getListByID(db, uint(listID), &list); err != nil {
		return fmt.Errorf("error fetching list: %v", err)
	}

	// Return the list's next page ID as JSON
	res := map[string]interface{}{
		"next_page_id": list.NextPageID,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return fmt.Errorf("error encoding JSON response: %v", err)
	}

	return nil
}

func handleGetPage(w http.ResponseWriter, r *http.Request) {
	if err := getPage(w, r); err != nil {
		log.Printf("Error in getPage: %v\n", err)
		if strings.Contains(err.Error(), "missing") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func getPage(w http.ResponseWriter, r *http.Request) error {
	// Extract the value of the "list_id" query parameter
	idStr := r.URL.Query().Get("page_id")
	if idStr == "" {
		return fmt.Errorf("page_id parameter is missing")
	}
	// Validate the list ID parameter
	pageID, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("page_id parameter is not a valid integer")
	}

	if db == nil { // check if db is nil
		return fmt.Errorf("database connection is nil")
	}

	var page Page
	// where "id" is the ID of the page you want to retrieve
	err = getPageByID(db, uint(pageID), &page)
	if err != nil {
		return fmt.Errorf("error getting page from database: %v", err)
	}

	// Query the database to get the articles associated with the specified page ID
	var articles []Article
	err = getArticlesByPageID(db, uint(pageID), &articles)
	if err != nil {
		return fmt.Errorf("error getting articles from database: %v", err)
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
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return fmt.Errorf("error encoding JSON response: %v", err)
	}

	return nil
}

func set(w http.ResponseWriter, r *http.Request) {
	var err error
	var article Article
	// Parse the request body to get the article data
	err = json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// log.Printf("Created new list: %+v\n", article)
	addArticleToPage(article)
}

func update(w http.ResponseWriter, r *http.Request) {
	var err error
	var id string

	vars := mux.Vars(r)
	id = vars["page_id"]

	pageID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	var articles []Article

	err = json.NewDecoder(r.Body).Decode(&articles)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Delete the existing articles associated with the page
	err = deleteArticlesByPageID(db, uint(pageID))
	if err != nil {
		http.Error(w, "Failed to delete articles", http.StatusInternalServerError)
		return
	}

	// Get the corresponding page
	var page Page
	err = getPageByID(db, uint(pageID), &page)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	// Update the page's articles
	page.Articles = articles

	err = savePage(db, &page)
	if err != nil {
		http.Error(w, "Failed to update page", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deletePage(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["list_id"]
	listID64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	listID := uint(listID64)

	err = deletePagesByListID(db, listID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, "Successfully deleted all pages and articles with list ID %d\n", listID)
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

	// err = preloadArticles(db, &page)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Printf("Articles in page %v\n", page.ID)
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
