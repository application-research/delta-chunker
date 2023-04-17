package model

import (
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenDatabase() (*gorm.DB, error) {

	dbName, okHost := viper.Get("DB_NAME").(string)
	if !okHost {
		panic("DB_NAME not set")
	}
	DB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})

	// generate new models.
	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	db.AutoMigrate(&Content{}, &ContentSplit{}, &ChunkTask{}, &ChunkRunConfig{})
}
