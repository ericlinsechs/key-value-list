package main

import "gorm.io/gorm"

// SaveArticle saves the given Article to the database.
func saveArticle(db *gorm.DB, article *Article) error {
	return db.Save(article).Error
}

func deleteArticlesByPageID(db *gorm.DB, pageID uint) error {
	err := db.Where("page_id = ?", pageID).Delete(&Article{}).Error
	if err != nil {
		return err
	}
	return nil
}
