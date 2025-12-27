package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&Entry{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	r := gin.Default()
	r.GET("/entry", getEntries(db))
	r.POST("/entry", postEntry(db))
	r.GET("/health", healthHandler(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
