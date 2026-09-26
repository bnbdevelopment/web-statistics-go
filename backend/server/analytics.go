package server

import (
	"log"
	"net/http"
	"statistics/analysis"
	"statistics/statistics"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func getFunnelStats(c *gin.Context) {
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"), 7)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	site := getSiteParam(c)
	stepsParam := c.Query("steps")
	var customSteps []string
	if stepsParam != "" {
		for _, s := range strings.Split(stepsParam, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				customSteps = append(customSteps, s)
			}
		}
	}
	result := statistics.GetFunnelStats(site, start, end, customSteps...)
	c.JSON(http.StatusOK, result)
}

func getEngagement(c *gin.Context) {
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"), 7)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	site := getSiteParam(c)
	sessions, sessionErr := analysis.GetSessionFeatures(site, start, end)
	if sessionErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": sessionErr.Error()})
		return
	}
	insights := analysis.GetEngagementInsights(sessions)
	c.JSON(http.StatusOK, gin.H{
		"segments":     insights.Segments,
		"topSessions":  insights.TopSessions,
		"averageScore": insights.Average,
	})
}

func getGeoTemporal(c *gin.Context) {
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"), 2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	site := getSiteParam(c)
	data, queryErr := analysis.GetGeoTemporalBuckets(site, start, end)
	if queryErr != nil {
		log.Println("GeoTemporal query error", queryErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load geo heatmap"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func getAlerts(c *gin.Context) {
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"), 7)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	site := getSiteParam(c)
	archetypes, archErr := analysis.GetArchetypes(site, start, end)
	if archErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": archErr.Error()})
		return
	}
	currentSessions, errSessions := analysis.GetSessionFeatures(site, start, end)
	if errSessions != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errSessions.Error()})
		return
	}
	prevStart := start.AddDate(0, 0, -7)
	prevSessions, _ := analysis.GetSessionFeatures(site, prevStart, start)
	currentEng := analysis.GetEngagementInsights(currentSessions)
	prevEng := analysis.GetEngagementInsights(prevSessions)
	bounce := statistics.GetBounceRate(start, end, site)
	alerts := analysis.GenerateAlerts(archetypes, currentEng.Average, prevEng.Average, bounce)
	c.JSON(http.StatusOK, alerts)
}

func getLandingPages(c *gin.Context) {
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"), 7)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	site := getSiteParam(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	results, queryErr := statistics.GetLandingPages(site, start, end, limit)
	if queryErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": queryErr.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func getExitPages(c *gin.Context) {
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"), 7)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	site := getSiteParam(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	results, queryErr := statistics.GetExitPages(site, start, end, limit)
	if queryErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": queryErr.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
