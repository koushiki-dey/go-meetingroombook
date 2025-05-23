package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var input models.Employee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	input.Password = string(hashedPassword)
	config.Connect()
	if err := config.GetDB().Create(&input).Error; err != nil {
		http.Error(w, "Failed to create employee", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input models.Employee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var employee models.Employee
	config.Connect()
	if err := config.GetDB().Where("email = ?", input.Email).First(&employee).Error; err != nil {
		http.Error(w, "Email not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(input.Password)); err != nil {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	session, _ := session.GetStore().Get(r, "session")
	session.Values["employee_id"] = employee.ID
	session.Save(r, w)

	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
