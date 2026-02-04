package statistics

import (
	"fmt"
	"math"
	"statistics/database"
	"statistics/structs"
	"time"
)

// Intermediate struct to hold raw query results for day of week analysis.
type dowResult struct {
	Day   int
	Count int
}

// countWeekdays counts the occurrences of each weekday within a given date range.
func countWeekdays(from, to time.Time) map[time.Weekday]int {
	counts := make(map[time.Weekday]int)
	// Normalize to the start of the day
	current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	endDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

	for !current.After(endDay) {
		counts[current.Weekday()]++
		current = current.AddDate(0, 0, 1)
	}
	return counts
}

// GetTrafficByDayOfWeek calculates the average traffic for each day of the week.
func GetTrafficByDayOfWeek(site string, from, to time.Time) ([]structs.TrafficByDay, error) {
	var results []dowResult
	var trafficByDay []structs.TrafficByDay

	dayMapping := []string{"Vasárnap", "Hétfő", "Kedd", "Szerda", "Csütörtök", "Péntek", "Szombat"}
	weekdayCounts := countWeekdays(from, to)

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

	dailyTotals := make(map[int]int)
	for _, res := range results {
		dailyTotals[res.Day] = res.Count
	}
	
	orderedDays := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}

	for _, day := range orderedDays {
		pgDow := int(day)
		totalCount := dailyTotals[pgDow]
		numOccurrences := weekdayCounts[day]
		
		var avg float64
		if numOccurrences > 0 {
			avg = float64(totalCount) / float64(numOccurrences)
		}

		trafficByDay = append(trafficByDay, structs.TrafficByDay{
			Day:   dayMapping[pgDow],
			Count: avg,
		})
	}

	return trafficByDay, nil
}

// GetTrafficByHourOfDay calculates the average traffic for each hour of the day.
func GetTrafficByHourOfDay(site string, from, to time.Time) ([]structs.TrafficByHour, error) {
	var results []structs.TrafficByHour

	// Calculate number of days in the range, rounding up. Minimum of 1.
	numberOfDays := math.Ceil(to.Sub(from).Hours() / 24)
	if numberOfDays < 1 {
		numberOfDays = 1
	}

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

	hourlyTotals := make(map[int]float64)
	for _, res := range results {
		hourlyTotals[res.Hour] = res.Count
	}

	var finalResults []structs.TrafficByHour
	for i := 0; i < 24; i++ {
		avgCount := float64(hourlyTotals[i]) / numberOfDays
		finalResults = append(finalResults, structs.TrafficByHour{
			Hour:  i,
			Count: avgCount,
		})
	}

	return finalResults, nil
}
