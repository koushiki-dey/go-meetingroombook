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

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	var room models.Room

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &room); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	config.Connect()
	db := config.GetDB()
	if err := db.Create(&room).Error; err != nil {
		http.Error(w, "Could not create room: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var createdRoom models.Room
	if err := db.First(&createdRoom, room.ID).Error; err != nil {
		http.Error(w, "Could not retrieve created room "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdRoom)
}

func GetRooms(w http.ResponseWriter, r *http.Request) {
	var rooms []models.Room
	config.Connect()
	db := config.GetDB()
	db.Preload("Bookings.Room").Preload("Bookings.Employee").Find(&rooms)

	resp, _ := json.Marshal(rooms)
	w.Header().Set("Content-type", "application/json")
	w.Write(resp)
}

func UpdateRoom(w http.ResponseWriter, r *http.Request) {
	var updateRoom = &models.Room{}
	utils.ParseBody(r, updateRoom)
	vars := mux.Vars(r)
	room_id := vars["id"]
	ID, err := strconv.ParseInt(room_id, 0, 0)
	if err != nil {
		fmt.Println("Error while parsing")
	}
	config.Connect()
	db := config.GetDB()
	var getRoom models.Room
	db = db.Where("ID=?", ID).Find(&getRoom)
	if updateRoom.Name != "" {
		getRoom.Name = updateRoom.Name
	}
	if updateRoom.Location != "" {
		getRoom.Location = updateRoom.Location
	}
	if updateRoom.Capacity != nil {
		getRoom.Capacity = updateRoom.Capacity
	}

	db.Save(&getRoom)
	res, _ := json.Marshal(&getRoom)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}
