package main

import (
	"log"
	"os"
	"statistics/database"
	"statistics/geolocation"
	"statistics/server"
)

func main() {
	// Database initialization
	error := database.DatabaseInitSession()
	if error != nil {
		panic("Failed to connect to the database: " + error.Error())
	} else {
		log.Println("Connected to TimescaleDB successfully")
	}

	// GeoIP initialization
	geoDBPath := os.Getenv("GEODB_PATH")
	if geoDBPath == "" {
		geoDBPath = "/geodb/GeoDB.mmdb" // Default path
	}

	if err := geolocation.InitGeoService(geoDBPath); err != nil {
		log.Printf("WARNING: Failed to initialize GeoIP service: %v", err)
		log.Println("Continuing without geolocation - geo fields will be null")
		// Don't panic - graceful degradation
	}
	defer geolocation.Close()

	server.Server()
}
