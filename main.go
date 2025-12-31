package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
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
	if err := db.AutoMigrate(&Entry{}, &Goal{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Ensure goal exists
	var goal Goal
	if err := db.First(&goal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default goal
			defaultGoal := Goal{Value: 100}
			if err := db.Create(&defaultGoal).Error; err != nil {
				log.Fatal("Failed to create default goal:", err)
			}
		} else {
			log.Fatal("Failed to check goal:", err)
		}
	}

	r := gin.Default()

	// Enable CORS for localhost and *.senthil.nz
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Allow localhost
			if origin == "http://localhost" || origin == "https://localhost" {
				return true
			}
			// Allow localhost with ports
			if len(origin) > 17 && origin[:17] == "http://localhost:" {
				return true
			}
			if len(origin) > 18 && origin[:18] == "https://localhost:" {
				return true
			}
			// Allow *.senthil.nz
			if len(origin) >= 10 && origin[len(origin)-10:] == ".senthil.nz" {
				return true
			}
			if origin == "http://senthil.nz" || origin == "https://senthil.nz" {
				return true
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key"},
		AllowCredentials: true,
	}))

	r.GET("/entry", getEntries(db))
	r.POST("/entry", postEntry(db))
	r.GET("/health", healthHandler(db))
	r.GET("/visits/progress/message", getProgressMessage(db))
	r.GET("/visits/streak", getStreak(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
