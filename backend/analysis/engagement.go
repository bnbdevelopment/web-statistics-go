package analysis

import (
	"sort"
	"statistics/structs"
)

// EngagementThresholds holds configurable buckets for segmentation.
type EngagementThresholds struct {
	HighlyEngaged float64
	Emerging      float64
}

var defaultThresholds = EngagementThresholds{
	HighlyEngaged: 75,
	Emerging:      40,
}

// GetEngagementInsights computes engagement segments and top sessions.
func GetEngagementInsights(features []SessionFeature) structs.EngagementResponse {
	if len(features) == 0 {
		return structs.EngagementResponse{
			Segments:    []structs.EngagementSegment{},
			TopSessions: []structs.EngagementLead{},
			Average:     0,
		}
	}

	thresholds := defaultThresholds

	totalScore := 0.0
	segmentCounts := map[string]int{
		"Highly Engaged": 0,
		"Emerging":       0,
		"At Risk":        0,
	}

	type scoredSession struct {
		SessionID string
		Score     float64
		Duration  float64
		Pages     int
	}

	var scored []scoredSession

	for _, feature := range features {
		score := feature.EngagementScore()
		totalScore += score

		switch {
		case score >= thresholds.HighlyEngaged:
			segmentCounts["Highly Engaged"]++
		case score >= thresholds.Emerging:
			segmentCounts["Emerging"]++
		default:
			segmentCounts["At Risk"]++
		}

		scored = append(scored, scoredSession{
			SessionID: feature.SessionID,
			Score:     score,
			Duration:  feature.Duration,
			Pages:     feature.PageCount,
		})
	}

	avgScore := totalScore / float64(len(features))

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	topLimit := 5
	if len(scored) < topLimit {
		topLimit = len(scored)
	}

	var topSessions []structs.EngagementLead
	for i := 0; i < topLimit; i++ {
		entry := scored[i]
		topSessions = append(topSessions, structs.EngagementLead{
			SessionID: entry.SessionID,
			Score:     entry.Score,
			Duration:  entry.Duration,
			Pages:     entry.Pages,
		})
	}

	segments := []structs.EngagementSegment{
		{
			Name:        "Highly Engaged",
			Description: "Score >= 75; fókuszált, magas értékű munkamenetek",
			Count:       segmentCounts["Highly Engaged"],
			Percentage:  percentage(segmentCounts["Highly Engaged"], len(features)),
		},
		{
			Name:        "Emerging",
			Description: "40-74 pont; potenciális növekedési lehetőség",
			Count:       segmentCounts["Emerging"],
			Percentage:  percentage(segmentCounts["Emerging"], len(features)),
		},
		{
			Name:        "At Risk",
			Description: "<40 pont; gyors kilépés vagy elakadás",
			Count:       segmentCounts["At Risk"],
			Percentage:  percentage(segmentCounts["At Risk"], len(features)),
		},
	}

	return structs.EngagementResponse{
		Segments:    segments,
		TopSessions: topSessions,
		Average:     avgScore,
	}
}

func percentage(count int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * 100
}
