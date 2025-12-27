package main

import "time"

type Entry struct {
	ID      uint      `json:"id" gorm:"primaryKey"`
	Date    time.Time `json:"date"`
	Visited bool      `json:"visited"`
}
