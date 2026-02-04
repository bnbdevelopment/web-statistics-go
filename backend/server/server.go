package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"statistics/database"
	"statistics/geolocation"
	"statistics/prometheus"
	"statistics/statistics"
	"statistics/structs"
	"time"

	gpmiddleware "github.com/carousell/gin-prometheus-middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func userTraffic(c *gin.Context) {
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		sessionId = uuid.New().String()
		c.String(http.StatusOK, sessionId)
		return
	} else {
		ip := c.Request.Header.Get("cf-connecting-ip")
		if ip == "" {
			ip = c.Request.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = c.ClientIP()
		}

		// Perform geolocation lookup
		geoData, _ := geolocation.Lookup(ip)

		record := structs.WebMetric{
			SessionId: sessionId,
			Timestamp: time.Now(),
			Page:      c.Query("page"),
			Site:      c.Query("site"),
			Ip:        ip,
		}

		// Populate geo fields if lookup succeeded
		if geoData != nil {
			record.CountryCode = &geoData.CountryCode
			record.CountryName = &geoData.CountryName
			record.City = &geoData.City
			record.Region = &geoData.Region
			record.Latitude = &geoData.Latitude
			record.Longitude = &geoData.Longitude
		}
		// If geoData is nil, fields remain nil (graceful degradation)

		err := database.Session.Create(&record).Error
		if err != nil {
			log.Println("Error inserting traffic data:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, sessionId)
	}
}

func getLocations(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	page := c.Query("page")
	var fromTime, toTime time.Time
	var err error
	layout := "2006-01-02"
	if !(from == "" || to == "") {
		fromTime, err = time.Parse(layout, from)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		toTime, err = time.Parse(layout, to)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		fromTime = time.Now().Add(-24 * time.Hour)
		toTime = time.Now()
	}
	locations := statistics.GetLocations(fromTime, toTime, page)
	c.JSON(http.StatusOK, gin.H{"locations": locations})
}

func traffic(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	page := c.Query("page")
	var fromTime, toTime time.Time
	var err error
	layout := "2006-01-02"
	if !(from == "" || to == "") {
		fromTime, err = time.Parse(layout, from)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		toTime, err = time.Parse(layout, to)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		fromTime = time.Now().Add(-24 * time.Hour)
		toTime = time.Now()
	}
	numberOfUsers := statistics.GetUsers(fromTime, toTime, page)
	c.JSON(http.StatusOK, gin.H{"traffic": numberOfUsers})
}

func getSites(c *gin.Context) {
	var sites []string
	err := database.Session.Model(&structs.WebMetric{}).Distinct("site").Pluck("site", &sites).Error
	if err != nil {
		log.Println("Error fetching distinct sites:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sites": sites})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, visitorkey")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func GetTimeOnTheSite(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	page := c.Query("page")
	layout := "2006-01-02"

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	start := end.Add(-24 * time.Hour)
	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	result := statistics.TimeOnSite(page, start, end)

	response := structs.AvgTimeResponse{AvgTimeSpent: result}

	c.JSON(http.StatusOK, response)
}

func getCohortData(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	site := c.Query("site")
	weeksStr := c.DefaultQuery("weeks", "12") // Default to 12 weeks
	layout := "2006-01-02"

	weeks, err := strconv.Atoi(weeksStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weeks format"})
		return
	}

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	start := end.AddDate(0, 0, -7*weeks) // Default start date based on number of weeks
	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	cohortData := statistics.GetCohortData(start, end, site, weeks)

	c.JSON(http.StatusOK, cohortData)
}

func getAverageJourney(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	site := c.Query("site")
	startPage := c.Query("start_page") // New parameter
	endPage := c.Query("end_page")     // New parameter
	layout := "2006-01-02"

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	start := end.Add(-24 * time.Hour) // Default to last 24 hours
	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	sankeyData := statistics.GetAverageJourney(start, end, site, startPage, endPage) // Pass new parameters
	c.JSON(http.StatusOK, sankeyData)
}

func getUniquePages(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	site := c.Query("site")
	layout := "2006-01-02"

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	start := end.Add(-24 * time.Hour) // Default to last 24 hours
	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	pages, err := statistics.GetAllUniquePages(site, start, end)
	if err != nil {
		log.Println("Error getting unique pages:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"pages": pages})
}

func getBounceRate(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	site := c.Query("site")
	layout := "2006-01-02"

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	start := end.Add(-24 * time.Hour)
	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	bounceRate := statistics.GetBounceRate(start, end, site)

	response := structs.BounceRateResponse{BounceRate: bounceRate}

	c.JSON(http.StatusOK, response)
}

func getTrafficByDayOfWeek(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	site := c.Query("site")
	layout := "2006-01-02"

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	start := end.AddDate(0, 0, -7) // Default to last 7 days
	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	traffic, err := statistics.GetTrafficByDayOfWeek(site, start, end)
	if err != nil {
		log.Println("Error getting traffic by day of week:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, traffic)
}

func getTrafficByHourOfDay(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	site := c.Query("site")
	layout := "2006-01-02"

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	start := end.Add(-24 * time.Hour) // Default to last 24 hours
	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	traffic, err := statistics.GetTrafficByHourOfDay(site, start, end)
	if err != nil {
		log.Println("Error getting traffic by hour of day:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, traffic)
}

func Server() {
	router := gin.Default()
	port := os.Getenv("BACKEND_PORT")
	prefix := os.Getenv("PREFIX")
	prometheus.RecordMetrics()
	p := gpmiddleware.NewPrometheus("gin")
	p.Use(router)

	if port == "" {
		port = "3001"
	}

	router.Use(CORSMiddleware())

	router.GET(prefix+"/put-traffic", userTraffic)

	router.POST(prefix+"/traffic", traffic)

	router.POST(prefix+"/sites", statistics.GetUsersByPages)

	router.POST(prefix+"/graph", statistics.GetTrafficStats)

	router.POST(prefix+"/active", statistics.GetActiveUsers)

	router.POST(prefix+"/time", statistics.GetTimeOnTheSite)

	router.POST(prefix+"/get-sites", getSites)

	router.POST(prefix+"/get-locations", getLocations)

	router.GET(prefix+"/bounce-rate", getBounceRate)

	router.POST(prefix+"/cohort", getCohortData)

	router.POST(prefix+"/average-journey", getAverageJourney)

	router.GET(prefix+"/statistics/traffic-by-day-of-week", getTrafficByDayOfWeek)
	router.GET(prefix+"/statistics/traffic-by-hour-of-day", getTrafficByHourOfDay)
	router.GET(prefix+"/statistics/unique-pages", getUniquePages)

	// Health check endpoint
	router.GET(prefix+"/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Println("prefix", prefix)
	log.Print("Starting server on port " + port)
	err := router.Run("0.0.0.0:" + port)
	if (err) == nil {
		log.Println("Failed to start server", "error", err)
		panic(err)
	}
}
