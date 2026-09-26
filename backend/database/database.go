package database

import (
	"fmt"
	"os"
	"statistics/structs"
	"time"

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

	// Configure connection pool
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	// Start asynchronous batch ingestion worker (10k buffer, 500 batch, 500ms flush)
	InitIngestionQueue(10000, 500, 500*time.Millisecond)

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

	if err := Session.AutoMigrate(&structs.WebMetric{}); err != nil {
		return err
	}

	// Performance indexes on web_metrics
	_ = Session.Exec("CREATE INDEX IF NOT EXISTS idx_web_metrics_site_ts ON web_metrics (site, timestamp DESC);")
	_ = Session.Exec("CREATE INDEX IF NOT EXISTS idx_web_metrics_session_ts ON web_metrics (session_id, timestamp ASC);")
	_ = Session.Exec("CREATE INDEX IF NOT EXISTS idx_web_metrics_site_page_ts ON web_metrics (site, page, timestamp);")
	_ = Session.Exec("CREATE INDEX IF NOT EXISTS idx_web_metrics_city_ts ON web_metrics (city, timestamp);")

	// Attempt TimescaleDB hypertable conversion if available
	_ = Session.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;")
	_ = Session.Exec("SELECT create_hypertable('web_metrics', 'timestamp', if_not_exists => TRUE);")

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
