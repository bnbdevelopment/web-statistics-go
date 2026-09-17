package server

import (
	"time"

	"github.com/gin-gonic/gin"
)

const dateLayout = "2006-01-02"

// parseDateRange returns defaulted date range if params empty and fallbackDays positive number of days.
func parseDateRange(startStr, endStr string, fallbackDays int) (time.Time, time.Time, error) {
	end := time.Now()
	var err error
	if endStr != "" {
		end, err = time.Parse(dateLayout, endStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	start := end.AddDate(0, 0, -fallbackDays)
	if startStr != "" {
		start, err = time.Parse(dateLayout, startStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return start, end, nil
}

// getSiteParam extracts the site query parameter with backward-compatibility fallback to 'page'.
func getSiteParam(c *gin.Context) string {
	site := c.Query("site")
	if site == "" {
		site = c.Query("page")
	}
	return site
}
