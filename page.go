package main

import "gorm.io/gorm"

func getPageByID(db *gorm.DB, id uint, page *Page) error {
	if err := db.Table("pages").First(page, id).Error; err != nil {
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

// CreatePage creates a new Page in the database.
func createPage(db *gorm.DB, page *Page) error {
	return db.Table("pages").Create(page).Error
}

// UpdateLastPageNextPageID updates the NextPageID of the last Page in the database.
func updateLastPageNextPageID(db *gorm.DB, lastPage *Page, newPageID uint) error {
	return db.Model(lastPage).UpdateColumn("next_page_id", newPageID).Error
}

// SavePage saves the given Page to the database.
func savePage(db *gorm.DB, page *Page) error {
	return db.Save(page).Error
}

func deletePagesByListID(db *gorm.DB, listID uint) error {
	// Delete all articles associated with the pages to be deleted
	if err := db.Where("page_id IN (SELECT id FROM pages WHERE list_id = ?)", listID).Delete(&Article{}).Error; err != nil {
		return err
	}

	// Delete all pages with the specified list ID
	if err := db.Where("list_id = ?", listID).Delete(&Page{}).Error; err != nil {
		return err
	}

	return nil
}
