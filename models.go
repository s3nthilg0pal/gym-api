package main

import "time"

type Entry struct {
	ID      uint      `json:"-" gorm:"primaryKey"` // Don't serialize ID in JSON
	Date    time.Time `json:"date"`
	Visited bool      `json:"visited"`
}

type EntryResponse struct {
	Date    time.Time `json:"date"`
	Visited bool      `json:"visited"`
}
