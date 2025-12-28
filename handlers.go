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

		// Check if entry already exists for this date
		var existing Entry
		if err := db.Where("date = ?", date).First(&existing).Error; err == nil {
			// Entry exists, return success (idempotent)
			c.JSON(http.StatusOK, gin.H{"message": "entry already exists"})
			return
		} else if err != gorm.ErrRecordNotFound {
			// Some other error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func getTotalVisits(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int64
		if err := db.Model(&Entry{}).Where("visited = ?", true).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var goal Goal
		if err := db.First(&goal).Error; err != nil {
			// If no goal exists, default to 100
			if err == gorm.ErrRecordNotFound {
				goal.Value = 100
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total_visits": count,
			"goal":         goal.Value,
		})
	}
}
