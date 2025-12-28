package prometheus

import (
	"fmt"
	"statistics/database"
	"statistics/statistics"
	"statistics/structs"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	visitorsBySite = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_traffic",
		Help: "Unique sessions (traffic) in last 24 hours by site",
	}, []string{"site"})
	activeUsersBySite = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_active_users",
		Help: "Number of currently active users by site (last 5 minutes)",
	}, []string{"site"})
	minuteSpentBySite = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_minute_spent",
		Help: "Average minutes spent on site in last 24h by site",
	}, []string{"site"})

	// Geolocation metrics
	trafficByCity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_traffic_by_city",
		Help: "Unique sessions in last 24 hours by city and country",
	}, []string{"site", "city", "country_code", "country_name"})
	activeUsersByCountry = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_active_users_by_country",
		Help: "Currently active users by country (last 5 minutes)",
	}, []string{"site", "country_code", "country_name"})
	trafficByCoordinates = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_traffic_coordinates",
		Help: "Traffic by geographic coordinates for geomap visualization",
	}, []string{"site", "latitude", "longitude", "city", "country_code"})
)

func RecordMetrics() {
	go func() {
		for {
			// Fetch distinct sites
			var sites []string
			_ = database.Session.Model(&structs.WebMetric{}).Distinct("site").Pluck("site", &sites).Error

			// Update metrics per site
			for _, site := range sites {
				if site == "" {
					continue
				}
				visitorsBySite.With(prometheus.Labels{"site": site}).Set(
					float64(statistics.GetUsers(time.Now().Add(-24*time.Hour), time.Now(), site)),
				)
				activeUsersBySite.With(prometheus.Labels{"site": site}).Set(
					float64(statistics.ActiveUsers(site)),
				)
				minuteSpentBySite.With(prometheus.Labels{"site": site}).Set(
					float64(statistics.TimeOnSite(site, time.Now().Add(-24*time.Hour), time.Now())),
				)

				// Update geolocation metrics
				updateGeoMetrics(site)
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

// updateGeoMetrics calculates and updates all geography-based metrics for a site
func updateGeoMetrics(site string) {
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	last5min := now.Add(-5 * time.Minute)

	// Traffic by city (last 24h)
	var cityStats []struct {
		City        string
		CountryCode string
		CountryName string
		Count       int64
	}

	query := `
		SELECT
			COALESCE(city, 'Unknown') as city,
			COALESCE(country_code, 'XX') as country_code,
			COALESCE(country_name, 'Unknown') as country_name,
			COUNT(DISTINCT session_id) as count
		FROM web_metrics
		WHERE timestamp >= ? AND timestamp <= ? AND site = ?
		GROUP BY city, country_code, country_name
	`
	_ = database.Session.Raw(query, last24h, now, site).Scan(&cityStats).Error

	for _, stat := range cityStats {
		trafficByCity.With(prometheus.Labels{
			"site":         site,
			"city":         stat.City,
			"country_code": stat.CountryCode,
			"country_name": stat.CountryName,
		}).Set(float64(stat.Count))
	}

	// Active users by country (last 5 min)
	var countryStats []struct {
		CountryCode string
		CountryName string
		Count       int64
	}

	countryQuery := `
		SELECT
			COALESCE(country_code, 'XX') as country_code,
			COALESCE(country_name, 'Unknown') as country_name,
			COUNT(DISTINCT session_id) as count
		FROM web_metrics
		WHERE timestamp >= ? AND timestamp <= ? AND site = ?
		GROUP BY country_code, country_name
	`
	_ = database.Session.Raw(countryQuery, last5min, now, site).Scan(&countryStats).Error

	for _, stat := range countryStats {
		activeUsersByCountry.With(prometheus.Labels{
			"site":         site,
			"country_code": stat.CountryCode,
			"country_name": stat.CountryName,
		}).Set(float64(stat.Count))
	}

	// Traffic by coordinates (for geomap)
	var coordStats []struct {
		Latitude    float64
		Longitude   float64
		City        string
		CountryCode string
		Count       int64
	}

	coordQuery := `
		SELECT
			latitude,
			longitude,
			COALESCE(city, 'Unknown') as city,
			COALESCE(country_code, 'XX') as country_code,
			COUNT(DISTINCT session_id) as count
		FROM web_metrics
		WHERE timestamp >= ? AND timestamp <= ? AND site = ?
		  AND latitude IS NOT NULL AND longitude IS NOT NULL
		GROUP BY latitude, longitude, city, country_code
	`
	_ = database.Session.Raw(coordQuery, last24h, now, site).Scan(&coordStats).Error

	for _, stat := range coordStats {
		trafficByCoordinates.With(prometheus.Labels{
			"site":         site,
			"latitude":     fmt.Sprintf("%.4f", stat.Latitude),
			"longitude":    fmt.Sprintf("%.4f", stat.Longitude),
			"city":         stat.City,
			"country_code": stat.CountryCode,
		}).Set(float64(stat.Count))
	}
}
