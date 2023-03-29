package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bytes"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	var err error
	// Initialize a new in-memory SQLite database
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Migrate the database schema
	db.AutoMigrate(&List{}, &Page{}, &Article{})

	createListIfNotExists()

	// Create same sample articles if there is no data in articles table
	// createSampleArticle()

	// Run the tests
	code := m.Run()

	// Exit with the appropriate code
	os.Exit(code)
}

func TestHandleGetHead(t *testing.T) {
	// Initialize the database and router
	router := mux.NewRouter()
	router.HandleFunc("/list/get", handleGetHead).Methods("GET")

	testCases := []struct {
		name         string
		url          string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Invalid list ID",
			url:          "/list/get?list_id=abc",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing list ID parameter",
			url:          "/list/get",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Valid list ID",
			url:          "/list/get?list_id=1",
			expectedCode: http.StatusOK,
			expectedBody: `{"next_page_id":1}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request
			req := httptest.NewRequest("GET", tc.url, nil)

			// Create a recorder to capture the response
			rec := httptest.NewRecorder()

			// Call the handler function
			router.ServeHTTP(rec, req)

			// Check the response status code
			if rec.Code != tc.expectedCode {
				t.Errorf("Expected status code %d but got %d", tc.expectedCode, rec.Code)
			}

			// Check the response body if it's expected
			if tc.expectedBody != "" && rec.Body.String() != tc.expectedBody+"\n" {
				t.Errorf("Handler returned unexpected body: got %q, want %q", rec.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestHandleGetPage(t *testing.T) {
	// Initialize test data
	pageID := 1
	articleTitle := "Test Article"
	articleAuthor := "Test Author"
	articleContent := "This is a test article."
	page := Page{
		ID:         uint(pageID),
		NextPageID: 2,
	}
	article := Article{
		Title:   articleTitle,
		Author:  articleAuthor,
		Content: articleContent,
		PageID:  uint(pageID),
	}

	db.Create(&page)
	db.Create(&article)

	// Initialize the database and router
	router := mux.NewRouter()
	router.HandleFunc("/page/get", handleGetPage).Methods("GET")

	// Create a request with an invalid page ID (not an integer)
	reqInvalidID := httptest.NewRequest("GET", "/page/get?page_id=abc", nil)

	// Create a recorder to capture the response
	rec := httptest.NewRecorder()

	// Test the function with an invalid page ID
	router.ServeHTTP(rec, reqInvalidID)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, rec.Code)
	}

	// Create a request without the page ID parameter
	reqMissingParam := httptest.NewRequest("GET", "/page/get", nil)

	// Create a recorder to capture the response
	rec = httptest.NewRecorder()

	// Test the function without the page ID parameter
	router.ServeHTTP(rec, reqMissingParam)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, rec.Code)
	}

	// Create a request with valid page ID parameter
	reqValidID := httptest.NewRequest("GET", "/page/get?page_id=1", nil)

	// Create a recorder to capture the response
	rec = httptest.NewRecorder()

	// Test the function with a valid page ID
	router.ServeHTTP(rec, reqValidID)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, rec.Code)
	}
	// Check the response body
	expected := `{"articles":[{"author":"Test Author","content":"This is a test article.","title":"Test Article"}],"next_page_id":2}` + "\n"
	if rec.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %q, want %q", rec.Body.String(), expected)
	}
}

func TestHandleSet(t *testing.T) {
	// Create a new request with a POST method and a JSON payload
	payload := []byte(`{"title": "Test Title", "author": "Test Author", "content": "Test Content"}`)
	req, err := http.NewRequest("POST", "/page/set", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Set the request header to indicate that the payload is in JSON format
	req.Header.Set("Content-Type", "application/json")

	// Create a new ResponseRecorder to record the HTTP response
	rr := httptest.NewRecorder()

	// Call the handleSet function and pass in the ResponseRecorder and the request
	handleSet(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the article was added to the page
	var page Page
	err = db.Preload("Articles").Last(&page).Error
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Articles) == 0 {
		t.Errorf("no article was added to the page")
	} else {
		addedArticle := page.Articles[len(page.Articles)-1]
		if addedArticle.Title != "Test Title" {
			t.Errorf("wrong article title: got %v want %v", addedArticle.Title, "Test Title")
		}
		if addedArticle.Author != "Test Author" {
			t.Errorf("wrong article author: got %v want %v", addedArticle.Author, "Test Author")
		}
		if addedArticle.Content != "Test Content" {
			t.Errorf("wrong article content: got %v want %v", addedArticle.Content, "Test Content")
		}
	}
}

func TestHandleUpdate(t *testing.T) {
	// Create a new test article
	article := Article{
		Title:   "Test Article",
		Author:  "Test Author",
		Content: "Test Content",
	}

	// Add the test article to a new test page
	addArticleToPage(article)

	// Get the ID of the test page
	var page Page
	err := getLastPage(db, &page)
	if err != nil {
		t.Errorf("Error getting last page: %v", err)
	}
	pageID := page.ID

	// Create a new test article with updated data
	updatedArticle := Article{
		Title:   "Updated Title",
		Author:  "Updated Author",
		Content: "Updated Content",
	}

	// Encode the updated article data as JSON
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(updatedArticle)
	if err != nil {
		t.Errorf("Error encoding JSON: %v", err)
	}

	// Create a new test HTTP request with the updated article data
	req, err := http.NewRequest("POST", fmt.Sprintf("/page/update?page_id=%d", pageID), &buf)
	if err != nil {
		t.Errorf("Error creating HTTP request: %v", err)
	}

	// Set the request header to indicate that the payload is in JSON format
	req.Header.Set("Content-Type", "application/json")

	// Create a new test HTTP response recorder
	rr := httptest.NewRecorder()

	// Call the handleUpdate function with the test HTTP request and response recorder
	handleUpdate(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Unexpected response status code: got %v, want %v", status, http.StatusOK)
	}

	// Get the page from the database
	var updatedPage Page
	err = db.Preload("Articles").First(&updatedPage, pageID).Error
	if err != nil {
		t.Errorf("Error getting updated page from database: %v", err)
	}

	// Check that the article on the page was updated with the new data
	if len(updatedPage.Articles) != 1 {
		t.Errorf("Unexpected number of articles on updated page: got %v, want %v", len(updatedPage.Articles), 1)
	} else {
		updatedArticle := updatedPage.Articles[0]
		if updatedArticle.Title != "Updated Title" || updatedArticle.Author != "Updated Author" || updatedArticle.Content != "Updated Content" {
			t.Errorf("Unexpected updated article data: got %+v, want %+v", updatedArticle, updatedArticle)
		}
	}
}

func TestHandleDeletePage(t *testing.T) {
	// Create a new test list
	list := List{ID: 99}

	// Save the test list to the database
	err := createList(db, &list)
	if err != nil {
		t.Errorf("Error saving list to database: %v", err)
	}

	// Create three test pages with the same list ID
	for i := 1; i <= 3; i++ {
		page := Page{ListID: list.ID}

		// Save the test page to the database
		err = savePage(db, &page)
		if err != nil {
			t.Errorf("Error saving page to database: %v", err)
		}

		// Add three test articles to the test page
		for j := 1; j <= 3; j++ {
			article := Article{
				Title:   fmt.Sprintf("Test Article %d-%d", i, j),
				Author:  "Test Author",
				Content: "Test Content",
				PageID:  page.ID,
			}

			// Save the test article to the database
			err = saveArticle(db, &article)
			if err != nil {
				t.Errorf("Error saving article to database: %v", err)
			}
		}
	}

	// Create a new test HTTP request to delete the pages with the list ID
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/page/delete?list_id=%d", list.ID), nil)
	if err != nil {
		t.Errorf("Error creating HTTP request: %v", err)
	}

	// Create a new test HTTP response recorder
	rr := httptest.NewRecorder()

	// Call the handleDeletePage function with the test HTTP request and response recorder
	handleDeletePage(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Unexpected response status code: got %v, want %v", status, http.StatusOK)
	}

	// Check that the pages and articles with the list ID were deleted from the database
	var pages []Page
	err = db.Where("list_id = ?", list.ID).Find(&pages).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Error getting pages from database: %v", err)
	} else if len(pages) != 0 {
		t.Errorf("Unexpected number of pages in database: got %v, want %v", len(pages), 0)
	}

	var articles []Article
	err = db.Where("page_id IN (SELECT id FROM pages WHERE list_id = ?)", list.ID).Find(&articles).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Error getting articles from database: %v", err)
	} else if len(articles) != 0 {
		t.Errorf("Unexpected number of articles in database: got %v, want %v", len(articles), 0)
	}
}
