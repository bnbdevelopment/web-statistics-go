package main

import (
	"statistics/database"
	"statistics/jobs"
	"statistics/server"
)

func main() {
	error := database.DatabaseInitSession()
	c := jobs.InitCronScheduler()
	if error != nil {
		panic("Failed to connect to the database: " + error.Error())
	} else {
		println("Connected to TimescaleDB successfully")
	}

	server.Server()

	defer c.Stop()

}
