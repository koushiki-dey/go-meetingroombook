package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// func CreateEmployee(w http.ResponseWriter, r *http.Request) {
// 	var emp models.Employee

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Failed to read request body", http.StatusBadRequest)
// 		return
// 	}

// 	if err := json.Unmarshal(body, &emp); err != nil {
// 		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	config.Connect()
// 	db := config.GetDB()
// 	if err := db.Create(&emp).Error; err != nil {
// 		http.Error(w, "Could not create employee: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	var createdEmployee models.Employee
// 	if err := db.First(&createdEmployee, emp.ID).Error; err != nil {
// 		http.Error(w, "Could not retrieve created employee: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(createdEmployee)
// }

func GetEmployees(w http.ResponseWriter, r *http.Request) {
	var employees []models.Employee
	config.Connect()
	db := config.GetDB()
	db.Preload("Bookings.Room").Preload("Bookings.Employee").Find(&employees)
	resp, _ := json.Marshal(employees)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func UpdateEmployees(w http.ResponseWriter, r *http.Request) {

	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid Employee ID", http.StatusBadRequest)
		return
	}

	var existing models.Employee
	config.Connect()
	db := config.GetDB()
	if err := db.First(&existing, id).Error; err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	if existing.ID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var updated models.Employee
	if err := json.Unmarshal(body, &updated); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	existing.Name = updated.Name
	existing.Email = updated.Email
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updated.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	existing.Password = string(hashedPassword)

	if err := db.Save(&existing).Error; err != nil {
		http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		return
	}
	resp, _ := json.Marshal(existing)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

}

func GetEmployee(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid Employee id", http.StatusBadRequest)
		return
	}

	var employee models.Employee
	config.Connect()
	db := config.GetDB()
	result := db.Preload("Bookings.Room").Preload("Bookings.Employee").First(&employee, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve employee", http.StatusInternalServerError)
		return
	}

	if employee.ID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	resp, _ := json.Marshal(employee)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
