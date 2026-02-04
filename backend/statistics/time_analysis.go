package statistics

import (
	"fmt"
	"statistics/database"
	"statistics/structs"
	"time"
)

// Intermediate struct to hold raw query results for day of week analysis.
type dowResult struct {
	Day   int
	Count int
}

// GetTrafficByDayOfWeek calculates the traffic for each day of the week within a given time range for a specific site.
func GetTrafficByDayOfWeek(site string, from, to time.Time) ([]structs.TrafficByDay, error) {
	var results []dowResult
	var trafficByDay []structs.TrafficByDay

	// Mapping from PostgreSQL DOW (0=Sun, 1=Mon, ...) to Hungarian day names.
	dayMapping := []string{"Vasárnap", "Hétfő", "Kedd", "Szerda", "Csütörtök", "Péntek", "Szombat"}

	query := database.Session.
		Model(&structs.WebMetric{}).
		Select("EXTRACT(DOW FROM timestamp) as day, COUNT(DISTINCT session_id) as count").
		Where("timestamp BETWEEN ? AND ?", from, to)

	if site != "" {
		query = query.Where("site = ?", site)
	}

	err := query.Group("day").Order("day").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query traffic by day of week: %w", err)
	}

	// Initialize map with all days to ensure the response contains all 7 days, even with 0 traffic
	dailyCounts := make(map[int]int)
	for i := 0; i < 7; i++ {
		dailyCounts[i] = 0
	}

	for _, res := range results {
		dailyCounts[res.Day] = res.Count
	}

	// Create the final ordered slice
	for i := 0; i < 7; i++ {
		dayIndex := (i + 1) % 7 // Start from Monday (1) instead of Sunday (0) for correct order
		if dayIndex == 0 {
			dayIndex = 7 // Adjust Sunday to be at the end of the week if we start from monday
		}
		pgDow := (dayIndex % 7) // Convert back to PG DOW index
		trafficByDay = append(trafficByDay, structs.TrafficByDay{
			Day:   dayMapping[pgDow],
			Count: dailyCounts[pgDow],
		})
	}
	// a hétfő legyen az első nap
	trafficByDay = append(trafficByDay[1:], trafficByDay[0])

	return trafficByDay, nil
}

// GetTrafficByHourOfDay calculates the traffic for each hour of the day within a given time range for a specific site.
func GetTrafficByHourOfDay(site string, from, to time.Time) ([]structs.TrafficByHour, error) {
	var results []structs.TrafficByHour

	query := database.Session.
		Model(&structs.WebMetric{}).
		Select("EXTRACT(HOUR FROM timestamp) as hour, COUNT(DISTINCT session_id) as count").
		Where("timestamp BETWEEN ? AND ?", from, to)

	if site != "" {
		query = query.Where("site = ?", site)
	}

	err := query.Group("hour").Order("hour").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query traffic by hour of day: %w", err)
	}

	// Ensure all 24 hours are present in the result
	hourlyCounts := make(map[int]int)
	for i := 0; i < 24; i++ {
		hourlyCounts[i] = 0
	}
	for _, res := range results {
		hourlyCounts[res.Hour] = res.Count
	}

	var finalResults []structs.TrafficByHour
	for i := 0; i < 24; i++ {
		finalResults = append(finalResults, structs.TrafficByHour{
			Hour:  i,
			Count: hourlyCounts[i],
		})
	}

	return finalResults, nil
}
