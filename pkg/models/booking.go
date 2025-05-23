package models

import (
	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model
	RoomID     uint   `json:"room_id"`
	EmployeeID uint   `json:"employee_id"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`

	Room     Room
	Employee Employee
}
