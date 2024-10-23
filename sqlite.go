package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

var (
	ErrOpenDB = errors.New("gorm open database error")
)

func RunSQLite() *gorm.DB {
	if db != nil {
		return db
	}

	file := filepath.Join(baseDir, "data.db")
	dsn := fmt.Sprintf("file:%s", file)

	var err error
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: NewGormLogger(log, 200*time.Millisecond)})
	if err != nil {
		log.Panicf("NewSQLite open database error: %s", err)
	}

	autoMigrate()

	log.Info("Started sqlite.")
	return db
}

func autoMigrate() {
	db.AutoMigrate(
		&TableColumn{},
		&TableCourse{},
	)
}


