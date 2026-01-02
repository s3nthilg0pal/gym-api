package main

import "time"

type Workout struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

type Entry struct {
	ID        uint      `json:"-" gorm:"primaryKey"`
	Date      time.Time `json:"date"`
	Visited   bool      `json:"visited"`
	WorkoutID *uint     `json:"workout_id,omitempty"`
	Workout   *Workout  `json:"workout,omitempty" gorm:"foreignKey:WorkoutID"`
}

type EntryResponse struct {
	Date    time.Time `json:"date"`
	Visited bool      `json:"visited"`
}

type Goal struct {
	ID         uint        `json:"-" gorm:"primaryKey"`
	Value      int         `json:"value"`
	Milestones []Milestone `json:"milestones,omitempty" gorm:"foreignKey:GoalID"`
}

type Milestone struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	GoalID uint   `json:"-"`
	Target int    `json:"target"`
	Name   string `json:"name"`
}
