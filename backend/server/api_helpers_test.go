package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestParseDateRange(t *testing.T) {
	// 1. Valid range
	startStr := "2026-09-01"
	endStr := "2026-09-15"
	start, end, err := parseDateRange(startStr, endStr, 7)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expectedStart, _ := time.Parse(dateLayout, startStr)
	expectedEnd, _ := time.Parse(dateLayout, endStr)
	if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
		t.Errorf("Date mismatch: got (%v, %v), want (%v, %v)", start, end, expectedStart, expectedEnd)
	}

	// 2. Empty inputs - should fallback
	fallbackDays := 14
	start, end, err = parseDateRange("", "", fallbackDays)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	diffDays := int(end.Sub(start).Hours() / 24)
	if diffDays < 13 || diffDays > 15 {
		t.Errorf("Expected approx %d days diff, got %d", fallbackDays, diffDays)
	}

	// 3. Invalid format
	_, _, err = parseDateRange("invalid-date", "", 7)
	if err == nil {
		t.Errorf("Expected error for invalid start date format")
	}

	_, _, err = parseDateRange("", "invalid-date", 7)
	if err == nil {
		t.Errorf("Expected error for invalid end date format")
	}
}

func TestGetSiteParam(t *testing.T) {
	// 1. 'site' parameter present
	req1 := httptest.NewRequest(http.MethodGet, "/test?site=mysite.hu&page=/contact", nil)
	c1, _ := gin.CreateTestContext(httptest.NewRecorder())
	c1.Request = req1
	if got := getSiteParam(c1); got != "mysite.hu" {
		t.Errorf("Expected mysite.hu, got %s", got)
	}

	// 2. Fallback to 'page' when 'site' is omitted
	req2 := httptest.NewRequest(http.MethodGet, "/test?page=legacy-site.hu", nil)
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = req2
	if got := getSiteParam(c2); got != "legacy-site.hu" {
		t.Errorf("Expected legacy-site.hu, got %s", got)
	}

	// 3. Both omitted
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	c3, _ := gin.CreateTestContext(httptest.NewRecorder())
	c3.Request = req3
	if got := getSiteParam(c3); got != "" {
		t.Errorf("Expected empty string, got %s", got)
	}
}
