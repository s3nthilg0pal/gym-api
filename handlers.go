package main

import (
	"fmt"
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
			Date:      date,
			Visited:   true,
			WorkoutID: nil,
		}

		if err := db.Create(&entry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "entry added"})
	}
}

func updateEntryWorkout(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			expectedKey = "default-secret"
		}
		if apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var payload struct {
			Date      string `json:"date"`
			WorkoutID uint   `json:"workout_id"`
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

		// Verify workout exists
		var workout Workout
		if err := db.First(&workout, payload.WorkoutID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "workout not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Find entry for the given date
		var entry Entry
		if err := db.Where("date = ?", date).First(&entry).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "no entry found for this date"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check if workout is already set
		if entry.WorkoutID != nil {
			c.JSON(http.StatusOK, gin.H{"message": "workout already set for this entry"})
			return
		}

		// Update the workout_id
		if err := db.Model(&entry).Update("workout_id", payload.WorkoutID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "entry updated with workout"})
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

func getProgressMessage(db *gorm.DB) gin.HandlerFunc {
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

		percent := 0
		if goal.Value > 0 {
			percent = int(float64(count) / float64(goal.Value) * 100)
		}

		var message string
		if percent >= 100 {
			message = fmt.Sprintf("🏆 Champion! You crushed it — %d of %d days!", count, goal.Value)
		} else if percent >= 80 {
			message = fmt.Sprintf("🔥 Almost there! %d of %d days - finish strong!", count, goal.Value)
		} else if percent >= 50 {
			message = fmt.Sprintf("💪 In the zone! %d of %d days - keep the momentum!", count, goal.Value)
		} else if percent >= 20 {
			message = fmt.Sprintf("🚀 Building habits! %d of %d days - you're on your way!", count, goal.Value)
		} else {
			message = fmt.Sprintf("🌱 Every rep counts! %d of %d days - let's go!", count, goal.Value)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": message,
		})
	}
}

func getStreak(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all entries ordered by date descending
		var entries []Entry
		if err := db.Where("visited = ?", true).Order("date DESC").Find(&entries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(entries) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"emoji":   "🎯",
				"tooltip": "Ready to begin? Your streak starts today!",
			})
			return
		}

		// Calculate streak
		today := time.Now().Truncate(24 * time.Hour)
		lastVisit := entries[0].Date.Truncate(24 * time.Hour)
		daysSinceLastVisit := int(today.Sub(lastVisit).Hours() / 24)

		// Count consecutive days from most recent visit
		streak := 1
		for i := 1; i < len(entries); i++ {
			expected := entries[i-1].Date.Truncate(24*time.Hour).AddDate(0, 0, -1)
			actual := entries[i].Date.Truncate(24 * time.Hour)
			if expected.Equal(actual) {
				streak++
			} else {
				break
			}
		}

		var emoji, tooltip string

		if daysSinceLastVisit > 1 {
			// Streak is broken - last visit was more than 1 day ago
			emoji = "💪"
			if streak > 1 {
				tooltip = fmt.Sprintf("Your %d day streak ended. Champions bounce back!", streak)
			} else {
				tooltip = "Time for a fresh start. Let's build a new streak!"
			}
		} else if streak >= 7 {
			// Epic streak
			emoji = "👑"
			tooltip = fmt.Sprintf("%d day streak! You're a legend!", streak)
		} else if streak >= 4 {
			// Solid streak
			emoji = "🔥"
			tooltip = fmt.Sprintf("%d day streak! You're on fire!", streak)
		} else {
			// Streak just started (1-3 days)
			emoji = "🌱"
			tooltip = fmt.Sprintf("%d day streak! Momentum is building!", streak)
		}

		c.JSON(http.StatusOK, gin.H{
			"emoji":   emoji,
			"tooltip": tooltip,
		})
	}
}

func getStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get total visits
		var totalVisits int64
		if err := db.Model(&Entry{}).Where("visited = ?", true).Count(&totalVisits).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get goal
		var goal Goal
		if err := db.First(&goal).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				goal.Value = 100
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Calculate progress percentage
		progress := 0
		if goal.Value > 0 {
			progress = int(float64(totalVisits) / float64(goal.Value) * 100)
		}

		// Get all entries ordered by date for streak calculation
		var entries []Entry
		if err := db.Where("visited = ?", true).Order("date DESC").Find(&entries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		currentStreak := 0
		longestStreak := 0

		if len(entries) > 0 {
			// Calculate current streak
			today := time.Now().Truncate(24 * time.Hour)
			lastVisit := entries[0].Date.Truncate(24 * time.Hour)
			daysSinceLastVisit := int(today.Sub(lastVisit).Hours() / 24)

			if daysSinceLastVisit <= 1 {
				// Streak is active
				currentStreak = 1
				for i := 1; i < len(entries); i++ {
					expected := entries[i-1].Date.Truncate(24*time.Hour).AddDate(0, 0, -1)
					actual := entries[i].Date.Truncate(24 * time.Hour)
					if expected.Equal(actual) {
						currentStreak++
					} else {
						break
					}
				}
			}

			// Calculate longest streak by checking all entries
			// Sort entries by date ascending for easier calculation
			var entriesAsc []Entry
			if err := db.Where("visited = ?", true).Order("date ASC").Find(&entriesAsc).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if len(entriesAsc) > 0 {
				tempStreak := 1
				longestStreak = 1
				for i := 1; i < len(entriesAsc); i++ {
					prevDate := entriesAsc[i-1].Date.Truncate(24 * time.Hour)
					currDate := entriesAsc[i].Date.Truncate(24 * time.Hour)
					expectedNext := prevDate.AddDate(0, 0, 1)
					if expectedNext.Equal(currDate) {
						tempStreak++
					} else {
						if tempStreak > longestStreak {
							longestStreak = tempStreak
						}
						tempStreak = 1
					}
				}
				if tempStreak > longestStreak {
					longestStreak = tempStreak
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"goal":          goal.Value,
			"progress":      progress,
			"currentStreak": fmt.Sprintf("%d days", currentStreak),
			"longestStreak": fmt.Sprintf("%d days", longestStreak),
		})
	}
}

func getWeeklyStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Calculate start of current week (Monday)
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday becomes 7
		}
		startOfWeek := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
		endOfWeek := startOfWeek.AddDate(0, 0, 7)

		// Count workouts completed this week
		var workoutsCompleted int64
		if err := db.Model(&Entry{}).Where("visited = ? AND date >= ? AND date < ?", true, startOfWeek, endOfWeek).Count(&workoutsCompleted).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Weekly goal (hardcoded for now)
		weeklyGoal := 5

		// Calculate progress percentage
		percent := 0
		if weeklyGoal > 0 {
			percent = int(float64(workoutsCompleted) / float64(weeklyGoal) * 100)
		}

		// Determine motivational message based on progress
		var message string
		if percent >= 100 {
			message = "🏆 Week conquered! You crushed your goal!"
		} else if percent >= 80 {
			message = "🔥 Almost there! Finish the week strong!"
		} else if percent >= 60 {
			message = "💪 Solid progress! Keep pushing!"
		} else if percent >= 40 {
			message = "🚀 Building momentum! You've got this!"
		} else if percent >= 20 {
			message = "🌱 Every rep counts! Keep showing up!"
		} else {
			message = "✨ Fresh week, fresh start! Today's your day!"
		}

		c.JSON(http.StatusOK, gin.H{
			"workouts_completed": workoutsCompleted,
			"weekly_goal":        weeklyGoal,
			"progress_message":   message,
		})
	}
}

func getMilestoneProgress(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get total visits
		var totalVisits int64
		if err := db.Model(&Entry{}).Where("visited = ?", true).Count(&totalVisits).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get all milestones ordered by target
		var milestones []Milestone
		if err := db.Order("target ASC").Find(&milestones).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(milestones) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message":        "🎯 No milestones set yet!",
				"total_workouts": totalVisits,
			})
			return
		}

		// Find the next milestone
		var nextMilestone *Milestone
		var lastCompletedMilestone *Milestone
		for i := range milestones {
			if int64(milestones[i].Target) > totalVisits {
				nextMilestone = &milestones[i]
				break
			}
			lastCompletedMilestone = &milestones[i]
		}

		var message string
		var milestoneName string
		var remaining int64

		if nextMilestone == nil {
			// All milestones completed!
			message = fmt.Sprintf("🏆 You've conquered all milestones! %d workouts and counting - you're a legend!", totalVisits)
			milestoneName = "All Complete!"
			remaining = 0
		} else {
			remaining = int64(nextMilestone.Target) - totalVisits
			milestoneName = nextMilestone.Name

			if remaining == 1 {
				message = fmt.Sprintf("⚡ ONE workout away from '%s'! This is your moment!", nextMilestone.Name)
			} else if remaining <= 3 {
				message = fmt.Sprintf("🔥 SO close to '%s'! Only %d workouts to go!", nextMilestone.Name, remaining)
			} else if remaining <= 5 {
				message = fmt.Sprintf("💪 '%s' is within reach! Just %d workouts away!", nextMilestone.Name, remaining)
			} else if lastCompletedMilestone != nil {
				message = fmt.Sprintf("🚀 '%s' unlocked! Next up: '%s' - %d workouts to go!", lastCompletedMilestone.Name, nextMilestone.Name, remaining)
			} else {
				message = fmt.Sprintf("🌱 Your journey begins! '%s' awaits - %d workouts to go!", nextMilestone.Name, remaining)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        message,
			"total_workouts": totalVisits,
			"next_milestone": milestoneName,
			"workouts_to_go": remaining,
		})
	}
}

func getForecast(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get total visits
		var totalVisits int64
		if err := db.Model(&Entry{}).Where("visited = ?", true).Count(&totalVisits).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get goal
		var goal Goal
		if err := db.First(&goal).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				goal.Value = 100
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Get first entry date to calculate weeks elapsed
		var firstEntry Entry
		if err := db.Where("visited = ?", true).Order("date ASC").First(&firstEntry).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{
					"current_progress": "No workouts yet - start your journey today!",
					"future_forecast":  "Complete your first workout to see your forecast.",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Calculate weeks elapsed since first workout
		now := time.Now()
		daysElapsed := now.Sub(firstEntry.Date).Hours() / 24
		weeksElapsed := daysElapsed / 7
		if weeksElapsed < 1 {
			weeksElapsed = 1 // Minimum 1 week to avoid division issues
		}

		// Calculate average workouts per week
		avgPerWeek := float64(totalVisits) / weeksElapsed

		// Format current progress message
		var currentProgress string
		if avgPerWeek >= 5 {
			currentProgress = fmt.Sprintf("🔥 You're crushing it with %.1f workouts per week!", avgPerWeek)
		} else if avgPerWeek >= 3 {
			currentProgress = fmt.Sprintf("💪 Solid pace! You're averaging %.1f workouts per week.", avgPerWeek)
		} else if avgPerWeek >= 1 {
			currentProgress = fmt.Sprintf("🌱 You're averaging %.1f workouts per week.", avgPerWeek)
		} else {
			currentProgress = fmt.Sprintf("📊 You're averaging %.1f workouts per week.", avgPerWeek)
		}

		// Calculate forecast
		var futureForecast string
		if totalVisits >= int64(goal.Value) {
			futureForecast = "🏆 You've already hit your goal! Keep the momentum going!"
		} else if avgPerWeek > 0 {
			remainingWorkouts := int64(goal.Value) - totalVisits
			weeksToGoal := float64(remainingWorkouts) / avgPerWeek
			completionDate := now.AddDate(0, 0, int(weeksToGoal*7))
			futureForecast = fmt.Sprintf("📅 At this pace, you'll hit your goal of %d by %s!", goal.Value, completionDate.Format("January 2, 2006"))
		} else {
			futureForecast = "Keep working out to see your forecast!"
		}

		c.JSON(http.StatusOK, gin.H{
			"current_progress": currentProgress,
			"future_forecast":  futureForecast,
		})
	}
}
