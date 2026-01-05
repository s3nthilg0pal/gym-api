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
	if err := db.AutoMigrate(&Workout{}, &Entry{}, &Goal{}, &Milestone{}); err != nil {
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
			goal = defaultGoal
		} else {
			log.Fatal("Failed to check goal:", err)
		}
	}

	// Seed milestones if none exist
	var milestoneCount int64
	db.Model(&Milestone{}).Count(&milestoneCount)
	if milestoneCount == 0 {
		milestones := []Milestone{
			{GoalID: goal.ID, Target: 15, Name: "Getting Started"},
			{GoalID: goal.ID, Target: 30, Name: "Building Habits"},
			{GoalID: goal.ID, Target: 50, Name: "Halfway Hero"},
			{GoalID: goal.ID, Target: 75, Name: "On Fire"},
			{GoalID: goal.ID, Target: 100, Name: "Goal Crusher"},
		}
		if err := db.Create(&milestones).Error; err != nil {
			log.Fatal("Failed to seed milestones:", err)
		}
	}

	// Seed workouts if none exist
	var workoutCount int64
	db.Model(&Workout{}).Count(&workoutCount)
	if workoutCount == 0 {
		workouts := []Workout{
			{Name: "Push"},
			{Name: "Pull"},
			{Name: "Legs"},
			{Name: "Cardio"},
		}
		if err := db.Create(&workouts).Error; err != nil {
			log.Fatal("Failed to seed workouts:", err)
		}
	}

	r := gin.Default()

	// Enable CORS for localhost and gym.senthil.nz
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
			// Allow gym.senthil.nz
			if origin == "http://gym.senthil.nz" || origin == "https://gym.senthil.nz" {
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
	r.PUT("/entry/workout", updateEntryWorkout(db))
	r.GET("/health", healthHandler(db))
	r.GET("/visits/progress/message", getProgressMessage(db))
	r.GET("/visits/streak", getStreak(db))
	r.GET("/visits/stats", getStats(db))
	r.GET("/visits/weekly", getWeeklyStats(db))
	r.GET("/visits/milestone", getMilestoneProgress(db))
	r.GET("/visits/forecast", getForecast(db))
	r.GET("/visits/ai-stats", getAIStats(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
