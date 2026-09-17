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
	if err := Connect(); err != nil {
		return err
	}

	return Migrate()
}

// Connect initializes the database session without changing the schema.
func Connect() error {
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
	return nil
}

// Migrate applies the schema changes required by the current backend version.
// It is intentionally separate from Connect so deployments can run it once
// before application containers start.
func Migrate() error {
	if Session == nil {
		return fmt.Errorf("database session is not initialized")
	}

	return Session.AutoMigrate(&structs.WebMetric{})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
