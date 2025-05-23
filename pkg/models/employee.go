package models

import (
	"gorm.io/gorm"
)

type Employee struct {
	gorm.Model
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Bookings []Booking `json:"bookings,omitempty"`
}
