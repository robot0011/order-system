package database

import (
	"log"
	"order-system/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/joho/godotenv"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	// Load environment variables from .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var dbURL = os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// Auto create tables
	err = DB.AutoMigrate(&models.User{}, &models.Restaurant{}, &models.Table{}, &models.MenuItem{}, &models.Order{}, &models.OrderItem{}, &models.Payment{})

	if err != nil {
		panic("Failed to migrate database!")
	}

}
