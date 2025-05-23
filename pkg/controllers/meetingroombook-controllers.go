package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
	"github.com/koushikidey/go-meetingroombook/pkg/utils"
	"gorm.io/gorm"
)

func CreateBooking(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	config.Connect()
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	var booking models.Booking
	if err := json.Unmarshal(body, &booking); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if booking.EndTime < booking.StartTime {
		http.Error(w, "End time is before start time", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateTimeFormat(booking.StartTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateTimeFormat(booking.EndTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	config.Connect()
	db := config.GetDB()
	var existingBookings []models.Booking
	db.Where("room_id = ?", booking.RoomID).Find(&existingBookings)

	var room models.Room
	db.Where("ID=?", booking.RoomID).Find(&room)
	currentCapacity := len(existingBookings) + 1
	maxCapacity := *room.Capacity
	_, err = utils.IsCapacityExceeding(currentCapacity, maxCapacity)
	if err != nil {
		http.Error(w, "Capacity Exceeded", http.StatusBadRequest)
		return
	}

	for _, b := range existingBookings {
		conflict, err := utils.IsTimeConflict(booking.StartTime, booking.EndTime, b.StartTime, b.EndTime)
		if err != nil {
			http.Error(w, "Error checking for conflicts", http.StatusInternalServerError)
			return
		}
		if conflict {
			http.Error(w, "Booking time conflicts with an existing booking", http.StatusConflict)
			return
		}
	}

	booking.EmployeeID = employeeID
	if err := db.Create(&booking).Error; err != nil {
		http.Error(w, "Could not create booking", http.StatusInternalServerError)
		return
	}

	var employee models.Employee
	db.First(&employee, employeeID)
	message := fmt.Sprintf("Hi %s,\n\nYour meeting room booking is confirmed from %s to %s in Room ID %d.",
		employee.Name, booking.StartTime, booking.EndTime, booking.RoomID)
	go utils.SendEmail(employee.Email, "Meeting Room Booking Confirmation", message)

	resp, _ := json.Marshal(booking)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func GetBookings(w http.ResponseWriter, r *http.Request) {
	// session, _ := session.GetStore().Get(r, "session")
	// employeeID, ok := session.Values["employee_id"].(uint)
	// if !ok {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	var bookings []models.Booking
	config.Connect()
	db := config.GetDB()
	//db.Preload("Room").Preload("Employee").Where("employee_id = ?", employeeID).Find(&bookings)
	db.Preload("Room").Preload("Employee").Find(&bookings)
	resp, _ := json.Marshal(bookings)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func GetBooking(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var booking models.Booking
	config.Connect()
	db := config.GetDB()
	result := db.Preload("Room").Preload("Employee").First(&booking, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Booking not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve booking", http.StatusInternalServerError)
		return
	}

	if booking.EmployeeID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	resp, _ := json.Marshal(booking)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func UpdateBooking(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var existing models.Booking
	config.Connect()
	db := config.GetDB()
	if err := db.First(&existing, id).Error; err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	if existing.EmployeeID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var updated models.Booking
	if err := json.Unmarshal(body, &updated); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateTimeFormat(updated.StartTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateTimeFormat(updated.EndTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var conflicts []models.Booking
	db.Where("room_id = ? AND id != ?", updated.RoomID, id).Find(&conflicts)
	for _, b := range conflicts {
		conflict, err := utils.IsTimeConflict(updated.StartTime, updated.EndTime, b.StartTime, b.EndTime)
		if err != nil {
			http.Error(w, "Error checking for conflicts", http.StatusInternalServerError)
			return
		}
		if conflict {
			http.Error(w, "Updated time conflicts with another booking", http.StatusConflict)
			return
		}
	}

	existing.RoomID = updated.RoomID
	existing.EmployeeID = updated.EmployeeID
	existing.StartTime = updated.StartTime
	existing.EndTime = updated.EndTime

	if err := db.Save(&existing).Error; err != nil {
		http.Error(w, "Failed to update booking", http.StatusInternalServerError)
		return
	}

	resp, _ := json.Marshal(existing)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func DeleteBooking(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}
	config.Connect()
	db := config.GetDB()
	var booking models.Booking
	if err := db.First(&booking, id).Error; err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	if booking.EmployeeID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := db.Delete(&booking).Error; err != nil {
		http.Error(w, "Failed to delete booking", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
