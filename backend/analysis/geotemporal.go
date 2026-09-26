package analysis

import (
	"statistics/database"
	"statistics/structs"
	"time"
)

// GetGeoTemporalBuckets aggregates session counts per city and hour.
func GetGeoTemporalBuckets(site string, from, to time.Time) ([]structs.GeoTemporalBucket, error) {
	query := database.Session.
		Model(&structs.WebMetric{}).
		Select("COALESCE(city, 'Ismeretlen') as city, COALESCE(latitude, 0) as latitude, COALESCE(longitude, 0) as longitude, EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as count").
		Where("timestamp BETWEEN ? AND ?", from, to)

	if site != "" {
		query = query.Where("site = ?", site)
	}

	var results []structs.GeoTemporalBucket

	if err := query.Group("city, latitude, longitude, hour").Order("count DESC").Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
