package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

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

func getPage(w http.ResponseWriter, r *http.Request) error {
	// Extract the value of the "page_id" query parameter
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

func set(w http.ResponseWriter, r *http.Request) error {
	var err error
	var article Article
	// Parse the request body to get the article data
	err = json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		return fmt.Errorf("invalid request body: %v", err)
	}

	addArticleToPage(article)

	w.WriteHeader(http.StatusOK)
	return nil
}

func update(w http.ResponseWriter, r *http.Request) error {
	// Extract the value of the "page_id" query parameter
	idStr := r.URL.Query().Get("page_id")
	if idStr == "" {
		return fmt.Errorf("page_id parameter is missing")
	}
	// Validate the list ID parameter
	pageID, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("page_id parameter is not a valid integer")
	}

	// Read the request body into a byte buffer
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	// Attempt to decode the request body as a slice of articles
	var articles []Article
	if err := json.Unmarshal(buf.Bytes(), &articles); err == nil {
		// Delete the existing articles associated with the page
		if err := deleteArticlesByPageID(db, uint(pageID)); err != nil {
			return fmt.Errorf("failed to delete articles: %v", err)
		}

		// Get the corresponding page
		var page Page
		if err := getPageByID(db, uint(pageID), &page); err != nil {
			return fmt.Errorf("page not found: %v", err)
		}

		// Update the page's articles
		page.Articles = articles

		if err := savePage(db, &page); err != nil {
			return fmt.Errorf("failed to update page: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}

	// Attempt to decode the request body as a single article
	var article Article
	if err := json.Unmarshal(buf.Bytes(), &article); err != nil {
		return fmt.Errorf("invalid request body: %v", err)
	}

	// Delete the existing articles associated with the page
	if err := deleteArticlesByPageID(db, uint(pageID)); err != nil {
		return fmt.Errorf("failed to delete articles: %v", err)
	}

	// Get the corresponding page
	var page Page
	if err := getPageByID(db, uint(pageID), &page); err != nil {
		return fmt.Errorf("page not found: %v", err)
	}

	// Update the page's articles
	page.Articles = []Article{article}

	if err := savePage(db, &page); err != nil {
		return fmt.Errorf("failed to update page: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func deletePage(w http.ResponseWriter, r *http.Request) error {
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

	err = deletePagesByListID(db, listID)
	if err != nil {
		return fmt.Errorf("failed to delete articles: %v", err)
	}

	fmt.Fprintf(w, "Successfully deleted all pages and articles with list ID %d\n", listID)
	return nil
}
