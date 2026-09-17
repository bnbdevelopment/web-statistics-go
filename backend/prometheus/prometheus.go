package prometheus

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"statistics/analysis"
	"statistics/database"
	"statistics/statistics"
	"statistics/structs"

	"github.com/mmcloughlin/geohash"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Core Traffic & Session KPIs (last 24 hours)
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

	totalRequestsBySite = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_total_requests",
		Help: "Total page requests in last 24 hours by site",
	}, []string{"site"})

	bounceRateBySite = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_bounce_rate_percentage",
		Help: "Bounce rate percentage in last 24 hours by site (0-100)",
	}, []string{"site"})

	bouncedSessionsBySite = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_bounced_sessions",
		Help: "Number of single-page bounced sessions in last 24 hours by site",
	}, []string{"site"})

	// Top Pages & Page Views
	pageViews = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_page_views",
		Help: "Total requests per page in last 24 hours by site and page",
	}, []string{"site", "page"})

	// Landing Pages Performance
	landingPageSessions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_landing_page_sessions",
		Help: "Sessions initiated on this landing page in last 24 hours",
	}, []string{"site", "page"})

	landingPageBounces = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_landing_page_bounces",
		Help: "Bounced sessions that landed on this page in last 24 hours",
	}, []string{"site", "page"})

	landingPageBounceRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_landing_page_bounce_rate",
		Help: "Bounce rate percentage for landing page in last 24 hours",
	}, []string{"site", "page"})

	// Exit Pages Performance (Churn Points)
	exitPageExits = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_exit_page_exits",
		Help: "Sessions ended on this exit page in last 24 hours",
	}, []string{"site", "page"})

	exitPageExitRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_exit_page_exit_rate",
		Help: "Exit rate percentage for page in last 24 hours",
	}, []string{"site", "page"})

	// Temporal Patterns
	trafficByHour = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_traffic_by_hour",
		Help: "Traffic distribution by hour of day (0-23) in last 24 hours",
	}, []string{"site", "hour"})

	trafficByWeekday = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_traffic_by_day_of_week",
		Help: "Average traffic by weekday over the last 7 days",
	}, []string{"site", "weekday"})

	// Visitor Archetypes & Behavioral Segmentation
	archetypeSessions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_archetype_sessions",
		Help: "Number of sessions categorized into archetype in last 24 hours",
	}, []string{"site", "archetype"})

	archetypePercentage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_archetype_percentage",
		Help: "Percentage of sessions categorized into archetype in last 24 hours",
	}, []string{"site", "archetype"})

	// Engagement Insights
	engagementSegmentSessions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_engagement_segment_sessions",
		Help: "Number of sessions in engagement segment (Highly Engaged, Emerging, At Risk)",
	}, []string{"site", "segment"})

	engagementAverageScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_engagement_average_score",
		Help: "Average engagement score (0-100) in last 24 hours by site",
	}, []string{"site"})

	// System Alerts
	alertActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statistics_alert_active",
		Help: "Active heuristic alert (1=active, 0=inactive)",
	}, []string{"site", "metric", "level"})

	// Geolocation Metrics
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
	}, []string{"site", "geohash", "latitude", "longitude", "city", "country_code"})

	// Ingestion Queue & Pipeline Health
	ingestionQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_ingestion_queue_depth",
		Help: "Current number of items in ingestion ring buffer",
	})

	ingestionQueueCapacity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_ingestion_queue_capacity",
		Help: "Total capacity of ingestion ring buffer",
	})

	ingestionTotalEvents = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_ingestion_events_total",
		Help: "Total telemetry events received by ingestion queue",
	})

	ingestionFallbackInserts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_ingestion_fallback_inserts_total",
		Help: "Total fallback synchronous database inserts when buffer full",
	})

	ingestionBatchErrors = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_ingestion_batch_errors_total",
		Help: "Total batch insert errors encountered by ingestion worker",
	})
)

// RecordMetrics starts the background loop exporting all system statistics to Prometheus.
func RecordMetrics() {
	go func() {
		// Run once on startup after a short grace period
		time.Sleep(2 * time.Second)
		collectAllMetrics()

		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			collectAllMetrics()
		}
	}()
}

// collectAllMetrics updates all Prometheus metrics across all sites and features.
func collectAllMetrics() {
	if database.Session == nil {
		return
	}

	// 1. Update Ingestion Pipeline Health
	if database.GlobalIngestionQueue != nil {
		qLen, qCap, totalIngested, fallbacks, batchErrors := database.GlobalIngestionQueue.GetStats()
		ingestionQueueDepth.Set(float64(qLen))
		ingestionQueueCapacity.Set(float64(qCap))
		ingestionTotalEvents.Set(float64(totalIngested))
		ingestionFallbackInserts.Set(float64(fallbacks))
		ingestionBatchErrors.Set(float64(batchErrors))
	}

	// 2. Fetch all active sites
	var sites []string
	if err := database.Session.Model(&structs.WebMetric{}).Distinct("site").Pluck("site", &sites).Error; err != nil {
		log.Printf("Prometheus metrics: failed to query distinct sites: %v", err)
		return
	}

	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	last7d := now.Add(-7 * 24 * time.Hour)
	prev24h := last24h.Add(-24 * time.Hour)

	for _, site := range sites {
		if site == "" {
			continue
		}

		// --- A. Core KPIs ---
		uniqueVisitors := statistics.GetUsers(last24h, now, site)
		visitorsBySite.With(prometheus.Labels{"site": site}).Set(float64(uniqueVisitors))

		activeUsers := statistics.ActiveUsers(site)
		activeUsersBySite.With(prometheus.Labels{"site": site}).Set(float64(activeUsers))

		timeSpent := statistics.TimeOnSite(site, last24h, now)
		minuteSpentBySite.With(prometheus.Labels{"site": site}).Set(timeSpent)

		var totalRequests int64
		_ = database.Session.Model(&structs.WebMetric{}).
			Where("site = ? AND timestamp >= ? AND timestamp <= ?", site, last24h, now).
			Count(&totalRequests).Error
		totalRequestsBySite.With(prometheus.Labels{"site": site}).Set(float64(totalRequests))

		bounceRate := statistics.GetBounceRate(last24h, now, site)
		bounceRateBySite.With(prometheus.Labels{"site": site}).Set(bounceRate)
		bouncedSessions := float64(uniqueVisitors) * (bounceRate / 100.0)
		bouncedSessionsBySite.With(prometheus.Labels{"site": site}).Set(bouncedSessions)

		// --- B. Top Page Views ---
		var pageStats []statistics.SiteTraffic
		pageQuery := `
			SELECT page, COUNT(*) AS count
			FROM (
				SELECT DISTINCT session_id, page
				FROM web_metrics
				WHERE timestamp >= ? AND timestamp <= ? AND site = ?
			) AS t
			GROUP BY page
			ORDER BY count DESC
			LIMIT 20;
		`
		if err := database.Session.Raw(pageQuery, last24h, now, site).Scan(&pageStats).Error; err == nil {
			for _, pageStat := range pageStats {
				pageViews.With(prometheus.Labels{
					"site": site,
					"page": pageStat.Page,
				}).Set(float64(pageStat.Count))
			}
		}

		// --- C. Landing Pages ---
		landingPages, landErr := statistics.GetLandingPages(site, last24h, now, 15)
		if landErr == nil {
			for _, landing := range landingPages {
				landingPageSessions.With(prometheus.Labels{
					"site": site,
					"page": landing.Page,
				}).Set(float64(landing.Sessions))

				landingPageBounces.With(prometheus.Labels{
					"site": site,
					"page": landing.Page,
				}).Set(float64(landing.Bounces))

				landingPageBounceRate.With(prometheus.Labels{
					"site": site,
					"page": landing.Page,
				}).Set(landing.BounceRate)
			}
		}

		// --- D. Exit Pages ---
		exitPages, exitErr := statistics.GetExitPages(site, last24h, now, 15)
		if exitErr == nil {
			for _, exit := range exitPages {
				exitPageExits.With(prometheus.Labels{
					"site": site,
					"page": exit.Page,
				}).Set(float64(exit.Exits))

				exitPageExitRate.With(prometheus.Labels{
					"site": site,
					"page": exit.Page,
				}).Set(exit.ExitRate)
			}
		}

		// --- E. Temporal Patterns (Hours & Weekdays) ---
		if hourly, err := statistics.GetTrafficByHourOfDay(site, last24h, now); err == nil {
			for _, h := range hourly {
				trafficByHour.With(prometheus.Labels{
					"site": site,
					"hour": strconv.Itoa(h.Hour),
				}).Set(float64(h.Count))
			}
		}

		if weekly, err := statistics.GetTrafficByDayOfWeek(site, last7d, now); err == nil {
			for _, w := range weekly {
				trafficByWeekday.With(prometheus.Labels{
					"site":    site,
					"weekday": w.Day,
				}).Set(w.Count)
			}
		}

		// --- F. Archetypes, Engagement & Alerts ---
		features, err := analysis.GetSessionFeatures(site, last24h, now)
		if err == nil && len(features) > 0 {
			// Archetypes
			archetypes, archErr := analysis.GetArchetypes(site, last24h, now)
			if archErr == nil {
				for _, arch := range archetypes {
					archSessions := float64(uniqueVisitors) * (arch.Percentage / 100.0)
					archetypeSessions.With(prometheus.Labels{
						"site":      site,
						"archetype": arch.Name,
					}).Set(archSessions)

					archetypePercentage.With(prometheus.Labels{
						"site":      site,
						"archetype": arch.Name,
					}).Set(arch.Percentage)
				}
			}

			// Engagement
			engagement := analysis.GetEngagementInsights(features)
			engagementAverageScore.With(prometheus.Labels{"site": site}).Set(engagement.Average)
			for _, seg := range engagement.Segments {
				engagementSegmentSessions.With(prometheus.Labels{
					"site":    site,
					"segment": seg.Name,
				}).Set(float64(seg.Count))
			}

			// Alerts
			prevFeatures, _ := analysis.GetSessionFeatures(site, prev24h, last24h)
			prevEngagement := analysis.GetEngagementInsights(prevFeatures)
			alerts := analysis.GenerateAlerts(archetypes, engagement.Average, prevEngagement.Average, bounceRate)

			// Record active alerts
			for _, alert := range alerts {
				alertActive.With(prometheus.Labels{
					"site":   site,
					"metric": alert.Metric,
					"level":  alert.Level,
				}).Set(1)
			}
		}

		// --- G. Geolocation Metrics ---
		updateGeoMetrics(site, now, last24h)
	}
}

// updateGeoMetrics calculates and updates all geography-based metrics for a site
func updateGeoMetrics(site string, now, last24h time.Time) {
	last5min := now.Add(-5 * time.Minute)

	// Traffic by city (last 24h, top 25)
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
		ORDER BY count DESC
		LIMIT 25
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

	// Traffic by coordinates (for geomap, top 50)
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
		ORDER BY count DESC
		LIMIT 50
	`
	_ = database.Session.Raw(coordQuery, last24h, now, site).Scan(&coordStats).Error

	for _, stat := range coordStats {
		hash := geohash.EncodeWithPrecision(stat.Latitude, stat.Longitude, 7)
		trafficByCoordinates.With(prometheus.Labels{
			"site":         site,
			"geohash":      hash,
			"latitude":     fmt.Sprintf("%.4f", stat.Latitude),
			"longitude":    fmt.Sprintf("%.4f", stat.Longitude),
			"city":         stat.City,
			"country_code": stat.CountryCode,
		}).Set(float64(stat.Count))
	}
}
