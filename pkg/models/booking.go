package models

import (
	"time"

	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model
	RoomID       uint      `json:"room_id"`
	EmployeeID   uint      `json:"employee_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	NumAttendees int       `json:"num_attendees"`
	ReminderSent bool
	CalendarID   string

	Room     Room
	Employee Employee
}
