package statistics

import (
	"fmt"
	"math"
	"net/http"
	"statistics/analysis"
	"statistics/config"
	"statistics/database"
	"statistics/structs"
	"strconv"
	"strings"
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
	var result structs.AvgTimeResponse // Using structs.AvgTimeResponse here
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
	site := c.Query("site")
	if site == "" {
		site = c.Query("page")
	}

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

	// Build base query
	var results []SiteTraffic
	if site == "" {
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
		if err := database.Session.Raw(query, start, end, site).Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, results)
}

type TrafficStat struct {
	Interval       int    `json:"interval"`
	Label          string `json:"label"`
	Timestamp      string `json:"timestamp"`
	UniqueSessions int    `json:"uniqueSessions"`
	TotalRequests  int    `json:"totalRequests"`
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
	site := c.Query("site")
	if site == "" {
		site = c.Query("page")
	}
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

	if site == "" {

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

		if err := database.Session.Raw(query, start, intervalDuration.Seconds(), start, end, site).Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	stats := make([]TrafficStat, intervals)
	for i := 0; i < intervals; i++ {
		tInterval := start.Add(time.Duration(i) * intervalDuration)
		var label string
		if totalDuration <= 36*time.Hour {
			label = tInterval.Format("15:04")
		} else {
			label = tInterval.Format("01-02 15:04")
		}

		stats[i] = TrafficStat{
			Interval:       i,
			Label:          label,
			Timestamp:      tInterval.Format(time.RFC3339),
			UniqueSessions: 0,
			TotalRequests:  0,
		}
	}

	for _, r := range results {
		if r.Interval >= 0 && r.Interval < intervals {
			stats[r.Interval].UniqueSessions = r.UniqueSessions
			stats[r.Interval].TotalRequests = r.TotalRequests
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

// GetFunnelStats aggregates how many unique sessions touch each funnel step in order.
func GetFunnelStats(site string, from, to time.Time, customSteps ...string) structs.FunnelResponse {
	var stepDefinitions []config.FunnelStep
	for _, s := range customSteps {
		s = strings.TrimSpace(s)
		if s != "" {
			stepDefinitions = append(stepDefinitions, config.FunnelStep{
				Label:    s,
				Keywords: []string{strings.ToLower(s)},
			})
		}
	}
	if len(stepDefinitions) == 0 {
		stepDefinitions = config.GetFunnelSteps(site)
	}
	if len(stepDefinitions) == 0 {
		return structs.FunnelResponse{Steps: []structs.FunnelStepStat{}}
	}

	sessionFeatures, err := analysis.GetSessionFeatures(site, from, to)
	if err != nil || len(sessionFeatures) == 0 {
		return structs.FunnelResponse{Steps: []structs.FunnelStepStat{}}
	}

	sessionResults := make(map[string][]bool)
	for _, feature := range sessionFeatures {
		reached := make([]bool, len(stepDefinitions))
		cursor := 0
		for stepIndex := 0; stepIndex < len(stepDefinitions); stepIndex++ {
			keywords := stepDefinitions[stepIndex].Keywords
			matched := false
			for cursor < len(feature.Pages) {
				page := strings.ToLower(feature.Pages[cursor])
				cursor++
				for _, keyword := range keywords {
					if keyword == "" {
						continue
					}
					if strings.Contains(page, keyword) {
						reached[stepIndex] = true
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				break
			}
		}
		sessionResults[feature.SessionID] = reached
	}

	counts := make([]int, len(stepDefinitions))
	firstStepSessions := 0
	for _, reached := range sessionResults {
		if len(reached) == 0 || !reached[0] {
			continue
		}
		firstStepSessions++
		for idx, ok := range reached {
			if ok {
				counts[idx]++
			} else {
				break
			}
		}
	}

	if firstStepSessions == 0 {
		return structs.FunnelResponse{Steps: []structs.FunnelStepStat{}}
	}

	var stepsStats []structs.FunnelStepStat
	prev := firstStepSessions
	for idx, step := range stepDefinitions {
		reached := counts[idx]
		conv := 0.0
		if prev > 0 {
			conv = math.Round((float64(reached)/float64(prev))*1000) / 10
		}
		drop := math.Max(0, 100-conv)
		stepsStats = append(stepsStats, structs.FunnelStepStat{
			Step:       step.Label,
			Users:      reached,
			Conversion: conv,
			Dropoff:    drop,
		})
		prev = reached
	}

	overall := 0.0
	if firstStepSessions > 0 {
		overall = math.Round((float64(prev)/float64(firstStepSessions))*1000) / 10
	}

	return structs.FunnelResponse{
		Steps:         stepsStats,
		FirstStepSize: firstStepSessions,
		Overall:       overall,
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

	result := TimeOnSite(page, start, end)

	response := AvgTimeResponse{AvgTimeSpent: result}

	c.JSON(http.StatusOK, response)
}

func GetBounceRate(start, end time.Time, site string) float64 {
	type bounceResult struct {
		TotalSessions   int64 `gorm:"column:total_sessions"`
		BouncedSessions int64 `gorm:"column:bounced_sessions"`
	}

	var res bounceResult
	var query string
	if site == "" {
		query = `
			SELECT 
				COUNT(*) AS total_sessions,
				COUNT(CASE WHEN page_count = 1 THEN 1 END) AS bounced_sessions
			FROM (
				SELECT session_id, COUNT(*) AS page_count
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ?
				GROUP BY session_id
			) s;
		`
		database.Session.Raw(query, start, end).Scan(&res)
	} else {
		query = `
			SELECT 
				COUNT(*) AS total_sessions,
				COUNT(CASE WHEN page_count = 1 THEN 1 END) AS bounced_sessions
			FROM (
				SELECT session_id, COUNT(*) AS page_count
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ? AND site = ?
				GROUP BY session_id
			) s;
		`
		database.Session.Raw(query, start, end, site).Scan(&res)
	}

	if res.TotalSessions == 0 {
		return 0.0
	}

	return math.Round((float64(res.BouncedSessions)/float64(res.TotalSessions)*100.0)*10) / 10
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

func GetAverageJourney(start, end time.Time, site, startPageFilter, endPageFilter string) structs.SankeyData {
	var flows []structs.FlowResult

	query := `
        WITH page_flows AS (
            SELECT
                page AS source_page,
                LEAD(page, 1) OVER (PARTITION BY session_id ORDER BY timestamp) AS target_page
            FROM
                web_metrics
            WHERE
                timestamp >= ? AND timestamp <= ?
                AND site LIKE ?
        )
        SELECT
            source_page,
            target_page,
            COUNT(*) AS flow_count
        FROM
            page_flows
        WHERE
            target_page IS NOT NULL AND source_page != target_page
    `
	if site == "" {
		site = "%"
	}
	queryParams := []interface{}{start, end, site}

	if startPageFilter != "" && startPageFilter != "%" {
		query += ` AND source_page = ?`
		queryParams = append(queryParams, startPageFilter)
	}
	if endPageFilter != "" && endPageFilter != "%" {
		query += ` AND target_page = ?`
		queryParams = append(queryParams, endPageFilter)
	}

	query += `
        GROUP BY
            source_page,
            target_page
        ORDER BY
            flow_count DESC
    `

	database.Session.Raw(query, queryParams...).Scan(&flows)

	// Process flows into Sankey data
	nodeMap := make(map[string]int)
	var nodes []structs.SankeyNode
	var links []structs.SankeyLink

	// Helper to add a node if it doesn't exist
	addNode := func(pageName string) {
		if _, exists := nodeMap[pageName]; !exists {
			nodeMap[pageName] = len(nodes)
			nodes = append(nodes, structs.SankeyNode{Name: pageName})
		}
	}

	for _, flow := range flows {
		addNode(flow.SourcePage)
		addNode(flow.TargetPage)

		links = append(links, structs.SankeyLink{
			Source: nodeMap[flow.SourcePage],
			Target: nodeMap[flow.TargetPage],
			Value:  flow.FlowCount,
		})
	}

	return structs.SankeyData{
		Nodes: nodes,
		Links: links,
	}
}

// GetAllUniquePages retrieves all unique page URLs within a given time range and site filter.
func GetAllUniquePages(site string, from, to time.Time) ([]string, error) {
	var pages []string

	query := database.Session.
		Model(&structs.WebMetric{}).
		Select("DISTINCT page").
		Where("timestamp >= ? AND timestamp <= ?", from, to)

	if site != "" {
		query = query.Where("site = ?", site)
	}

	err := query.Order("page ASC").Pluck("page", &pages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get unique pages: %w", err)
	}

	return pages, nil
}

// GetLandingPages returns the most common initial landing pages and their bounce rates.
func GetLandingPages(site string, from, to time.Time, limit int) ([]structs.LandingPageStat, error) {
	if limit <= 0 {
		limit = 10
	}
	var results []structs.LandingPageStat

	var query string
	var err error
	if site == "" {
		query = `
			WITH session_landing AS (
				SELECT 
					session_id,
					(ARRAY_AGG(page ORDER BY timestamp ASC))[1] AS landing_page,
					COUNT(*) AS total_views
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ?
				GROUP BY session_id
			)
			SELECT 
				landing_page AS page,
				COUNT(*) AS sessions,
				COUNT(CASE WHEN total_views = 1 THEN 1 END) AS bounces,
				ROUND((COUNT(CASE WHEN total_views = 1 THEN 1 END)::numeric / NULLIF(COUNT(*), 0)::numeric) * 100, 1)::float8 AS bounce_rate
			FROM session_landing
			WHERE landing_page IS NOT NULL AND landing_page != ''
			GROUP BY landing_page
			ORDER BY sessions DESC
			LIMIT ?;
		`
		err = database.Session.Raw(query, from, to, limit).Scan(&results).Error
	} else {
		query = `
			WITH session_landing AS (
				SELECT 
					session_id,
					(ARRAY_AGG(page ORDER BY timestamp ASC))[1] AS landing_page,
					COUNT(*) AS total_views
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ? AND site = ?
				GROUP BY session_id
			)
			SELECT 
				landing_page AS page,
				COUNT(*) AS sessions,
				COUNT(CASE WHEN total_views = 1 THEN 1 END) AS bounces,
				ROUND((COUNT(CASE WHEN total_views = 1 THEN 1 END)::numeric / NULLIF(COUNT(*), 0)::numeric) * 100, 1)::float8 AS bounce_rate
			FROM session_landing
			WHERE landing_page IS NOT NULL AND landing_page != ''
			GROUP BY landing_page
			ORDER BY sessions DESC
			LIMIT ?;
		`
		err = database.Session.Raw(query, from, to, site, limit).Scan(&results).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get landing pages: %w", err)
	}
	return results, nil
}

// GetExitPages returns the most common departure pages where sessions ended.
func GetExitPages(site string, from, to time.Time, limit int) ([]structs.ExitPageStat, error) {
	if limit <= 0 {
		limit = 10
	}
	var results []structs.ExitPageStat

	var query string
	var err error
	if site == "" {
		query = `
			WITH session_exit AS (
				SELECT 
					session_id,
					(ARRAY_AGG(page ORDER BY timestamp DESC))[1] AS exit_page
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ?
				GROUP BY session_id
			),
			total_count AS (
				SELECT COUNT(*) AS total_sessions FROM session_exit
			)
			SELECT 
				exit_page AS page,
				COUNT(*) AS exits,
				ROUND((COUNT(*)::numeric / NULLIF((SELECT total_sessions FROM total_count), 0)::numeric) * 100, 1)::float8 AS exit_rate
			FROM session_exit
			WHERE exit_page IS NOT NULL AND exit_page != ''
			GROUP BY exit_page
			ORDER BY exits DESC
			LIMIT ?;
		`
		err = database.Session.Raw(query, from, to, limit).Scan(&results).Error
	} else {
		query = `
			WITH session_exit AS (
				SELECT 
					session_id,
					(ARRAY_AGG(page ORDER BY timestamp DESC))[1] AS exit_page
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ? AND site = ?
				GROUP BY session_id
			),
			total_count AS (
				SELECT COUNT(*) AS total_sessions FROM session_exit
			)
			SELECT 
				exit_page AS page,
				COUNT(*) AS exits,
				ROUND((COUNT(*)::numeric / NULLIF((SELECT total_sessions FROM total_count), 0)::numeric) * 100, 1)::float8 AS exit_rate
			FROM session_exit
			WHERE exit_page IS NOT NULL AND exit_page != ''
			GROUP BY exit_page
			ORDER BY exits DESC
			LIMIT ?;
		`
		err = database.Session.Raw(query, from, to, site, limit).Scan(&results).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get exit pages: %w", err)
	}
	return results, nil
}
