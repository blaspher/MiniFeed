package config

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"minifeed/internal/model"
)

func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("connect mysql err: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Post{}, &model.Follow{}); err != nil {
		log.Fatalf("auto migrate err: %v", err)
	}

	return db

}
