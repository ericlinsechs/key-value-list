package main

import "gorm.io/gorm"

func getListByID(db *gorm.DB, id uint, list *List) error {
	if err := db.Table("lists").First(list, id).Error; err != nil {
		return err
	}
	return nil
}

// CreatePage creates a new Page in the database.
func createList(db *gorm.DB, list *List) error {
	return db.Table("lists").Create(list).Error
}
