package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getEntries(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entries []Entry
		if err := db.Find(&entries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, entries)
	}
}

func postEntry(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			expectedKey = "default-secret" // for development
		}
		if apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var payload struct {
			Date string `json:"date"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		date, err := time.Parse("2006-01-02", payload.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}

		entry := Entry{
			Date:    date,
			Visited: true,
		}

		if err := db.Create(&entry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "entry added"})
	}
}

func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connectivity
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database connection error"})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database ping failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	}
}
