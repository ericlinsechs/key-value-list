package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	var err error
	// Initialize a new in-memory SQLite database
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	// Migrate the database schema
	if err := db.AutoMigrate(&List{}, &Page{}, &Article{}); err != nil {
		fmt.Println("Error migrating schema:", err)
		os.Exit(1)
	}

	createListIfNotExists()

	// Create same sample articles if there is no data in articles table
	createSampleArticle()

	r := mux.NewRouter()

	// list
	r.HandleFunc("/list/get", handleGetHead).Methods("GET")

	// page
	r.HandleFunc("/page/get", handleGetPage).Methods("GET")

	// Run the tests
	code := m.Run()

	os.Exit(code)
}

func TestHandleGetHead(t *testing.T) {
	// r := mux.NewRouter()
	// r.HandleFunc("/list/get", handleGetHead).Methods("GET")

	// Call the getHead function with a valid list ID
	req, err := http.NewRequest("GET", "/list/get?list_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleGetHead)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handleGetHead returned wrong status code: got %v, expectedPage %v", status, http.StatusOK)
	}

	// Decode the JSON response into a map
	var res map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	// Check if the response contains the expectedPage data
	if nextPageID, ok := res["next_page_id"]; !ok || nextPageID != float64(1) {
		t.Errorf("handleGetHead returned unexpectedPage response: %v", res)
	}
}

func TestHandleGetPage(t *testing.T) {
	expectedPage := Page{
		Articles: []Article{
			{Title: "Article 1", Author: "Author 1", Content: "Content 1"},
			{Title: "Article 2", Author: "Author 2", Content: "Content 2"},
			{Title: "Article 3", Author: "Author 3", Content: "Content 3"},
			{Title: "Article 4", Author: "Author 4", Content: "Content 4"},
			{Title: "Article 5", Author: "Author 5", Content: "Content 5"},
		},
		NextPageID: 2,
	}

	// Create a new request with a "page_id" query parameter
	req, err := http.NewRequest("GET", "/page/get?page_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the handler function with the test request and response recorder
	handler := http.HandlerFunc(handleGetPage)
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response content type
	if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ctype, "application/json")
	}

	// Parse the response body into a Page struct
	var resp Page
	err = json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	// Check the articles
	if len(resp.Articles) != len(expectedPage.Articles) {
		t.Errorf("handler returned wrong number of articles: got %v want %v",
			len(resp.Articles), len(expectedPage.Articles))
	}
	for i, a := range resp.Articles {
		if a.Title != expectedPage.Articles[i].Title {
			t.Errorf("handler returned wrong article title: got %v want %v",
				a.Title, expectedPage.Articles[i].Title)
		}
		if a.Author != expectedPage.Articles[i].Author {
			t.Errorf("handler returned wrong article author: got %v want %v",
				a.Author, expectedPage.Articles[i].Author)
		}
		if a.Content != expectedPage.Articles[i].Content {
			t.Errorf("handler returned wrong article content: got %v want %v",
				a.Content, expectedPage.Articles[i].Content)
		}
	}

	// Check the next page ID
	if resp.NextPageID != expectedPage.NextPageID {
		t.Errorf("handler returned wrong next page ID: got %v want %v",
			resp.NextPageID, expectedPage.NextPageID)
	}
}
