package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/googleapi"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	"github.com/koushikidey/go-meetingroombook/pkg/routes"
	"github.com/koushikidey/go-meetingroombook/pkg/utils"
	"github.com/robfig/cron/v3"
)

var reminderCron *cron.Cron

func startReminderJob() {
	log.Println("startReminderJob() called")

	reminderCron = cron.New()

	_, err := reminderCron.AddFunc("@every 1m", func() {
		log.Println("Cron job triggered at", time.Now().Format(time.RFC3339))

		db := config.GetDB()
		var bookings []models.Booking
		now := time.Now().UTC()
		future := now.Add(10 * time.Minute)
		log.Printf("Looking for bookings between %s and %s",
			now.Format(time.RFC3339),
			future.Format(time.RFC3339),
		)
		result := db.Where("start_time BETWEEN ? AND ?", now, future).Find(&bookings)
		if result.Error != nil {
			log.Printf("Error fetching bookings: %v", result.Error)
			return
		}

		log.Printf("Found %d upcoming bookings", len(bookings))

		for _, booking := range bookings {
			var employee models.Employee
			db.First(&employee, booking.EmployeeID)

			msg := fmt.Sprintf("Reminder: You have a meeting in Room %d from %s to %s.",
				booking.RoomID,
				booking.StartTime.Format(time.RFC3339),
				booking.EndTime.Format(time.RFC3339),
			)

			log.Printf(" Sending reminder to %s", employee.Email)
			go utils.SendEmail(employee.Email, "Meeting Reminder", msg)
		}
	})

	if err != nil {
		log.Fatalf(" Failed to add cron job: %v", err)
	}

	reminderCron.Start()
}

func main() {
	log.Println("Application starting...")

	config.Connect()

	startReminderJob()
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	googleapi.InitOAuth(clientID, clientSecret, redirectURL)

	router := mux.NewRouter()
	routes.RegisterMeetingRoomRoutes(router)
	//http.Handle("/", router)
	//router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))
	//router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./frontend/"))))
	//router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("../../frontend/index.html"))))
	// Check if file exists
	// path := "../../frontend"
	// _, err := os.Stat(filepath.Join(path, "index.html"))
	// if err != nil {
	// 	log.Fatalf("index.html not found at %s: %v", path, err)
	// } else {
	// 	log.Println("index.html found and will be served from:", path)
	// }
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working dir:", err)
	}

	// This now points to your actual frontend folder
	frontendPath := filepath.Join(wd, "frontend")

	// Check if index.html exists
	_, err = os.Stat(filepath.Join(frontendPath, "index.html"))
	if err != nil {
		log.Fatalf("index.html not found at %s: %v", frontendPath, err)
	}

	log.Println("index.html found! Serving from:", frontendPath)

	// Serve files
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(frontendPath))))

	// Serve static
	//router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(path))))

	log.Println(" Server running at http://localhost:9010")
	log.Fatal(http.ListenAndServe("localhost:9010", router))
}
