package config

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"minifeed/internal/model"
)

func InitDB() *gorm.DB {
	dsn := "root:@LJYmm20040208@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"
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
