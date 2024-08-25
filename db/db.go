package db

import (
	"os"
	"os/user"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	path := "tstud.db"
	if os.Getenv("TSTUD_ENV") == "development" {
		path = filepath.Join("./", path)
	} else {
		path = filepath.Join(usr.HomeDir, path)
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&File{}, &Tag{}, &Alias{}); err != nil {
		return nil, err
	}

	return db, nil
}
