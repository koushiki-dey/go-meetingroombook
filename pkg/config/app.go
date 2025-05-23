package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

var (
	db *gorm.DB
)

func GetDB() *gorm.DB {
	return db
}

func MigrateDB(db *gorm.DB) {
	db.AutoMigrate(&models.Room{}, &models.Employee{}, &models.Booking{})
}
func Connect() {
	dsn := "root:@midnighTS13@tcp(127.0.0.1:3306)/simplerest?charset=utf8mb4&parseTime=True&loc=Local"

	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	db = d
	MigrateDB(db)
}
