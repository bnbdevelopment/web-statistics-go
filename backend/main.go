package main

import (
	"io"
	"log"
	"os"
	"statistics/database"
	"statistics/geolocation"
	"statistics/server"
)

func main() {
	closeLog := configureLogging()
	if closeLog != nil {
		defer closeLog()
	}

	if err := database.Connect(); err != nil {
		panic("Failed to connect to the database: " + err.Error())
	}

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := database.Migrate(); err != nil {
			panic("Failed to migrate the database: " + err.Error())
		}
		log.Println("Database migration completed successfully")
		return
	}

	log.Println("Connected to TimescaleDB successfully")

	// GeoIP initialization
	geoDBPath := os.Getenv("GEODB_PATH")
	if geoDBPath == "" {
		geoDBPath = "/geodb/GeoDB.mmdb" // Default path
	}

	if err := geolocation.InitGeoService(geoDBPath); err != nil {
		log.Printf("WARNING: Failed to initialize GeoIP service: %v", err)
		log.Println("Continuing without geolocation - geo fields will be null")
	}
	defer geolocation.Close()

	server.Server()
}

func configureLogging() func() {
	path := os.Getenv("LOG_FILE")
	if path == "" {
		return nil
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		log.Printf("WARNING: Failed to open log file %s: %v", path, err)
		return nil
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))
	return func() { _ = file.Close() }
}
