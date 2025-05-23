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
	"github.com/koushikidey/go-meetingroombook/pkg/utils"
)

func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var emp models.Employee

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &emp); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	config.Connect()
	db := config.GetDB()
	if err := db.Create(&emp).Error; err != nil {
		http.Error(w, "Could not create employee: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var createdEmployee models.Employee
	if err := db.First(&createdEmployee, emp.ID).Error; err != nil {
		http.Error(w, "Could not retrieve created employee: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdEmployee)
}

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
	var updateEmployee = &models.Employee{}
	utils.ParseBody(r, updateEmployee)
	vars := mux.Vars(r)
	employee_id := vars["id"]
	ID, err := strconv.ParseInt(employee_id, 0, 0)
	if err != nil {
		fmt.Println("Error while parsing")
	}
	config.Connect()
	db := config.GetDB()
	var getEmployee models.Employee
	db = db.Where("ID=?", ID).Find(&getEmployee)
	if updateEmployee.Name != "" {
		getEmployee.Name = updateEmployee.Name
	}
	if updateEmployee.Email != "" {
		getEmployee.Email = updateEmployee.Email
	}
	db.Save(&getEmployee)
	res, _ := json.Marshal(&getEmployee)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}
