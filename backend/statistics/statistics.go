package statistics

import (
	"net/http"
	"statistics/database"
	"statistics/structs"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUsers(t1 time.Time, t2 time.Time, site string) int {
	var results int

	if site == "" {
		query := `SELECT COUNT (*) from (SELECT session_id FROM "web_metrics" WHERE "timestamp" >= ? AND "timestamp" <= ? GROUP BY session_id) as lamdba;`
		database.Session.Raw(query, t1, t2).Scan(&results)
	}
	if site != "" {
		query := `SELECT COUNT (*) from (SELECT session_id FROM "web_metrics" WHERE "timestamp" >= ? AND "timestamp" <= ? AND site = ? GROUP BY session_id) as lamdba;`
		database.Session.Raw(query, t1, t2, site).Scan(&results)
	}

	return results
}

func GetLocations(t1 time.Time, t2 time.Time, site string) []structs.LocationQueryResult {
	var results []structs.LocationQueryResult
	if site == "" {
		query := `SELECT city, latitude, longitude, COUNT(DISTINCT session_id) as user_count FROM "web_metrics" WHERE "timestamp" >= ? AND "timestamp" <= ? AND city != '' GROUP BY city, latitude, longitude`
		database.Session.Raw(query, t1, t2).Scan(&results)
	} else {
		query := `SELECT city, latitude, longitude, COUNT(DISTINCT session_id) as user_count FROM "web_metrics" WHERE "timestamp" >= ? AND "timestamp" <= ? AND site = ? AND city != '' GROUP BY city, latitude, longitude`
		database.Session.Raw(query, t1, t2, site).Scan(&results)
	}
	return results
}

func ActiveUsers(page string) int64 {
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)

	var count int64

	if page == "" {
		if err := database.Session.
			Model(&structs.WebMetric{}).
			Where("timestamp >= ? AND timestamp <= ?", fiveMinutesAgo, now).
			Distinct("session_id").
			Count(&count).Error; err != nil {
			return 0
		}
	} else {
		if err := database.Session.
			Model(&structs.WebMetric{}).
			Where("timestamp >= ? AND timestamp <= ? AND site = ?", fiveMinutesAgo, now, page).
			Distinct("session_id").
			Count(&count).Error; err != nil {
			return 0
		}

	}
	return count
}

func TimeOnSite(page string, start time.Time, end time.Time) float64 {
	var result AvgTimeResponse
	if page == "" {
		query := `
			WITH diffs AS (
			SELECT
				session_id,
				EXTRACT(EPOCH FROM (timestamp - lag(timestamp) OVER (PARTITION BY session_id ORDER BY timestamp))) / 60.0 AS minutes_diff
			FROM web_metrics
			WHERE timestamp >= ? AND timestamp <= ?
		), session_times AS (
			SELECT
				session_id,
				SUM(CASE WHEN minutes_diff IS NOT NULL AND minutes_diff <= 5 THEN minutes_diff ELSE 0 END) AS total_time
			FROM diffs
			GROUP BY session_id
		)
		SELECT COALESCE(AVG(total_time), 0) AS avg_time_spent FROM session_times;
		`

		if err := database.Session.Raw(query, start, end).Scan(&result).Error; err != nil {
			return 0.0
		}
	} else {
		query := `
			WITH diffs AS (
			SELECT
				session_id,
				EXTRACT(EPOCH FROM (timestamp - lag(timestamp) OVER (PARTITION BY session_id ORDER BY timestamp))) / 60.0 AS minutes_diff
			FROM web_metrics
			WHERE timestamp >= ? AND timestamp <= ? AND site = ?
		), session_times AS (
			SELECT
				session_id,
				SUM(CASE WHEN minutes_diff IS NOT NULL AND minutes_diff <= 5 THEN minutes_diff ELSE 0 END) AS total_time
			FROM diffs
			GROUP BY session_id
		)
		SELECT COALESCE(AVG(total_time), 0) AS avg_time_spent FROM session_times;
		`
		if err := database.Session.Raw(query, start, end, page).Scan(&result).Error; err != nil {
			return 0.0
		}
	}

	return result.AvgTimeSpent
}

type SiteTraffic struct {
	Page  string `json:"page"`
	Count int    `json:"count"`
}

func GetUsersByPages(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	page := c.Query("page")

	end := time.Now()
	if endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		end = t
	}

	// default start date = 24h before end
	start := end.Add(-24 * time.Hour)
	if startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		start = t
	}

	// Build base query – you’ll need a table with at least session_id, url, time
	var results []SiteTraffic
	if page == "" {
		query := `
				SELECT page, COUNT(*) AS count
				FROM (
					SELECT DISTINCT session_id, page
					FROM web_metrics
					WHERE timestamp >= ? AND timestamp <= ?
				) AS t
				GROUP BY page
				ORDER BY count DESC;
			`
		if err := database.Session.Raw(query, start, end).Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		query := `
				SELECT page, COUNT(*) AS count
				FROM (
					SELECT DISTINCT session_id, page
					FROM web_metrics
					WHERE timestamp >= ? AND timestamp <= ? AND site = ?
				) AS t
				GROUP BY page
				ORDER BY count DESC;
			`
		if err := database.Session.Raw(query, start, end, page).Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, results)
}

type TrafficStat struct {
	Interval       int `json:"interval"`
	UniqueSessions int `json:"uniqueSessions"`
	TotalRequests  int `json:"totalRequests"`
}

// DB model for your traffic table
type Traffic struct {
	ID        uint      `gorm:"primaryKey"`
	SessionID string    `gorm:"column:session_id"`
	Time      time.Time `gorm:"column:time"`
}

func GetTrafficStats(c *gin.Context) {
	startStr := c.Query("from")
	endStr := c.Query("to")
	intervalsStr := c.DefaultQuery("intervals", "10")
	page := c.Query("page")
	layout := "2006-01-02"
	// default: last 24h
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

	intervals, err := strconv.Atoi(intervalsStr)
	if err != nil || intervals <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Intervals must be greater than 0"})
		return
	}

	totalDuration := end.Sub(start)
	intervalDuration := totalDuration / time.Duration(intervals)

	type Result struct {
		Interval       int
		UniqueSessions int
		TotalRequests  int
	}
	var results []Result

	if page == "" {

		query := `
			WITH interval_data AS (
				SELECT
					floor(extract(epoch from (timestamp - ?)) / ?)::int as interval,
					session_id,
					count(*) as cnt
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ?
				GROUP BY interval, session_id
			)
			SELECT
				interval,
				count(DISTINCT session_id) as unique_sessions,
				sum(cnt) as total_requests
			FROM interval_data
			GROUP BY interval
			ORDER BY interval
		`

		if err := database.Session.Raw(query, start, intervalDuration.Seconds(), start, end).Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		query := `
	WITH interval_data AS (
		SELECT
			floor(extract(epoch from (timestamp - ?)) / ?)::int as interval,
			session_id,
			count(*) as cnt
		FROM web_metrics
		WHERE timestamp >= ? AND timestamp <= ? AND site = ?
		GROUP BY interval, session_id
	)
	SELECT
		interval,
		count(DISTINCT session_id) as unique_sessions,
		sum(cnt) as total_requests
	FROM interval_data
	GROUP BY interval
	ORDER BY interval
`

		if err := database.Session.Raw(query, start, intervalDuration.Seconds(), start, end, page).Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	stats := make([]TrafficStat, intervals)
	for i := 0; i < intervals; i++ {
		stats[i] = TrafficStat{
			Interval:       i,
			UniqueSessions: 0,
			TotalRequests:  0,
		}
	}

	for _, r := range results {
		if r.Interval >= 0 && r.Interval < intervals {
			stats[r.Interval] = TrafficStat{
				Interval:       r.Interval,
				UniqueSessions: r.UniqueSessions,
				TotalRequests:  r.TotalRequests,
			}
		}
	}

	c.JSON(http.StatusOK, stats)
}

type ActiveUsersResponse struct {
	Count int `json:"count"`
}

func GetActiveUsers(c *gin.Context) {

	page := c.Query("page")

	count := ActiveUsers(page)

	c.JSON(http.StatusOK, ActiveUsersResponse{Count: int(count)})
}

type AvgTimeResponse struct {
	AvgTimeSpent float64 `json:"avgTimeSpent"`
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

	result := TimeOnSite(page, start, end)

	response := AvgTimeResponse{AvgTimeSpent: result}

	c.JSON(http.StatusOK, response)
}

func GetBounceRate(start, end time.Time, site string) float64 {
	var totalSessions int64
	var bouncedSessions int64

	// Query for total sessions within the time range and optionally filtered by site
	dbQuery := database.Session.
		Model(&structs.WebMetric{}).
		Where("timestamp >= ? AND timestamp <= ?", start, end)

	if site != "" {
		dbQuery = dbQuery.Where("site = ?", site)
	}

	dbQuery.Distinct("session_id").Count(&totalSessions)

	// Query for bounced sessions (sessions with only one page view)
	// Subquery to count entries per session_id within the time range and site filter
	subQuery := database.Session.
		Select("session_id").
		Table("web_metrics").
		Where("timestamp >= ? AND timestamp <= ?", start, end)

	if site != "" {
		subQuery = subQuery.Where("site = ?", site)
	}

	subQuery.
		Group("session_id").
		Having("COUNT(*) = 1")

	// Main query to count how many distinct session_ids from the subquery exist
	database.Session.
		Table("(?) as bounced", subQuery).
		Count(&bouncedSessions)

	if totalSessions == 0 {
		return 0.0
	}

	return float64(bouncedSessions) / float64(totalSessions) * 100.0
}

func GetCohortData(start, end time.Time, site string, numberOfWeeks int) []structs.CohortData {
	var results []structs.CohortRow

	query := `
        WITH user_first_visit AS (
            SELECT
                session_id,
                DATE_TRUNC('week', MIN(timestamp)) AS cohort_week
            FROM
                web_metrics
            WHERE
                site LIKE ? AND timestamp >= ? AND timestamp <= ?
            GROUP BY
                session_id
        ),
        weekly_activity AS (
            SELECT DISTINCT
                session_id,
                DATE_TRUNC('week', timestamp) AS activity_week
            FROM
                web_metrics
            WHERE
                site LIKE ? AND timestamp >= ? AND timestamp <= ?
        ),
        cohort_activity AS (
            SELECT
                ufv.cohort_week,
                TRUNC(EXTRACT(EPOCH FROM (wa.activity_week - ufv.cohort_week)) / (7 * 24 * 60 * 60)) AS week_number,
                COUNT(DISTINCT ufv.session_id) as user_count
            FROM
                user_first_visit ufv
            JOIN
                weekly_activity wa ON ufv.session_id = wa.session_id
            GROUP BY
                ufv.cohort_week,
                week_number
        )
        SELECT
            cohort_week,
            week_number,
            user_count
        FROM
            cohort_activity
        ORDER BY
            cohort_week DESC, week_number ASC
    `

	if site == "" {
		site = "%"
	}

	if err := database.Session.Raw(query, site, start, end, site, start, end).Scan(&results).Error; err != nil {
		// Handle error
		return nil
	}

	// Process the raw results into the final format
	cohortMap := make(map[time.Time]map[int]int)
	for _, row := range results {
		if _, ok := cohortMap[row.CohortWeek]; !ok {
			cohortMap[row.CohortWeek] = make(map[int]int)
		}
		cohortMap[row.CohortWeek][row.WeekNumber] = row.UserCount
	}

	var cohortData []structs.CohortData
	for cohortWeek, weekData := range cohortMap {
		totalUsers := weekData[0]
		if totalUsers == 0 {
			continue
		}

		retentionData := make([]float64, numberOfWeeks)
		for i := 0; i < numberOfWeeks; i++ {
			if users, ok := weekData[i]; ok {
				retentionData[i] = (float64(users) / float64(totalUsers)) * 100.0
			} else {
				retentionData[i] = 0.0
			}
		}

		cohortData = append(cohortData, structs.CohortData{
			CohortDate:    cohortWeek.Format("2006-01-02"),
			TotalUsers:    totalUsers,
			RetentionData: retentionData,
		})
	}

	return cohortData
}
