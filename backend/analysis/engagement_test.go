package analysis

import (
	"testing"
	"time"
)

func TestEngagementScore(t *testing.T) {
	now := time.Now()

	// 1. Bounced / short session
	bounced := SessionFeature{
		SessionID:       "s-bounced",
		Duration:        10,
		PageCount:       1,
		UniquePageCount: 1,
		LoopScore:       1.0,
		FirstEvent:      now,
		LastEvent:       now.Add(10 * time.Second),
		Pages:           []string{"/"},
	}
	score := bounced.EngagementScore()
	if score > 20 {
		t.Errorf("Expected low engagement score for bounced session, got %.2f", score)
	}

	// 2. High engagement session
	engaged := SessionFeature{
		SessionID:       "s-engaged",
		Duration:        450,
		PageCount:       6,
		UniquePageCount: 5,
		LoopScore:       1.2,
		FirstEvent:      now,
		LastEvent:       now.Add(450 * time.Second),
		Pages:           []string{"/", "/features", "/docs", "/pricing", "/signup", "/dashboard"},
	}
	engagedScore := engaged.EngagementScore()
	if engagedScore < 60 {
		t.Errorf("Expected high engagement score for multi-page active session, got %.2f", engagedScore)
	}
}

func TestGetEngagementInsights(t *testing.T) {
	features := []SessionFeature{
		{
			SessionID:       "s-1",
			Duration:        5,
			PageCount:       1,
			UniquePageCount: 1,
		},
		{
			SessionID:       "s-2",
			Duration:        500,
			PageCount:       8,
			UniquePageCount: 6,
		},
	}

	insights := GetEngagementInsights(features)
	if len(insights.Segments) != 3 {
		t.Fatalf("Expected 3 segments, got %d", len(insights.Segments))
	}
	if insights.Average <= 0 {
		t.Errorf("Expected positive average engagement score, got %.2f", insights.Average)
	}
}
