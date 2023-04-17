package model

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenDatabase(name string) (*gorm.DB, error) {

	DB, err := gorm.Open(sqlite.Open(name), &gorm.Config{})

	// generate new models.
	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	db.AutoMigrate(&Content{}, &ContentSplit{})
}
