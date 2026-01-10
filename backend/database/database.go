package database

import (
	"fmt"
	"os"
	"statistics/structs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Session *gorm.DB

func DatabaseInitSession() error {
	host := getEnv("DB_HOST", "timescaledb")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "12345")
	dbname := getEnv("DB_NAME", "statistics")
	port := getEnv("DB_PORT", "5432")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	Session = db

	err = db.AutoMigrate(&structs.WebMetric{})
	if err != nil {
		return err
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
