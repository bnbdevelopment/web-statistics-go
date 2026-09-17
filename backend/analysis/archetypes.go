package analysis

import (
	"fmt"
	"math"
	"sort"
	"statistics/database"
	"statistics/structs"
	"strings"
	"time"
)

const (
	ArchetypeTargeted   = "Célirányos Látogató"
	ArchetypeEngaged    = "Elmélyült Böngésző"
	ArchetypeFrustrated = "Frusztrált vagy Céltalan"
	ArchetypeDefault    = "Általános Látogató"
)

// SessionFeature holds the calculated behavioral metrics for a single session.
type SessionFeature struct {
	SessionID       string
	Duration        float64 // in seconds
	PageCount       int
	UniquePageCount int
	LoopScore       float64 // Ratio of total pages to unique pages
	FirstEvent      time.Time
	LastEvent       time.Time
	Pages           []string
}

// EngagementScore returns a 0-100 heuristic score with larger values for
// efficient, long sessions and lower values for short/bouncy ones.
func (s SessionFeature) EngagementScore() float64 {
	// Base score on unique pages with diminishing returns
	uniqueComponent := math.Min(float64(s.UniquePageCount)*12, 50)

	// Time component: full score reached around 10 minutes
	const maxDuration = 600.0
	durationComponent := math.Min((s.Duration/maxDuration)*30, 30)

	// Penalize looping/navigation inefficiency (higher loop score => lower score)
	loopPenalty := math.Max(0, (s.LoopScore-1.2)*15)

	// Penalize very short sessions heavily
	shortPenalty := 0.0
	if s.Duration < 30 || s.PageCount <= 1 {
		shortPenalty = 25
	}

	raw := uniqueComponent + durationComponent - loopPenalty - shortPenalty
	if raw < 0 {
		return 0
	}
	if raw > 100 {
		return 100
	}
	return math.Round(raw*10) / 10
}

// GetSessionFeatures queries and calculates behavioral features for all sessions in a given timeframe.
func GetSessionFeatures(site string, from, to time.Time) ([]SessionFeature, error) {
	type rawSessionData struct {
		SessionID       string
		StartTime       time.Time
		EndTime         time.Time
		PageCount       int
		UniquePageCount int
		Pages           string
	}

	var rawData []rawSessionData

	dbQuery := database.Session.
		Model(&structs.WebMetric{}).
		Select(`
	            session_id,
	            MIN(timestamp) as start_time,
	            MAX(timestamp) as end_time,
	            COUNT(*) as page_count,
	            COUNT(DISTINCT page) as unique_page_count,
	            STRING_AGG(COALESCE(page, '/'), '||' ORDER BY timestamp) as pages
	        `).
		Where("timestamp BETWEEN ? AND ?", from, to)

	if site != "" {
		dbQuery = dbQuery.Where("site = ?", site)
	}

	err := dbQuery.Group("session_id").Find(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query session data: %w", err)
	}

	var features []SessionFeature
	for _, data := range rawData {
		duration := data.EndTime.Sub(data.StartTime).Seconds()
		var loopScore float64
		if data.UniquePageCount > 0 {
			loopScore = float64(data.PageCount) / float64(data.UniquePageCount)
		}

		features = append(features, SessionFeature{
			SessionID:       data.SessionID,
			Duration:        duration,
			PageCount:       data.PageCount,
			UniquePageCount: data.UniquePageCount,
			LoopScore:       loopScore,
			FirstEvent:      data.StartTime,
			LastEvent:       data.EndTime,
			Pages:           strings.Split(data.Pages, "||"),
		})
	}
	return features, nil
}

// classifySession applies a set of rules to categorize a session into an archetype.
func classifySession(feature SessionFeature) string {
	// Rule for "Frustrated or Aimless" - requires high loop score AND significant duration
	if feature.LoopScore >= 2.5 && feature.PageCount > 5 && feature.Duration > 180 { // 180 seconds = 3 minutes
		return ArchetypeFrustrated
	}
	// Another potential rule for frustrated/aimless
	if feature.LoopScore >= 1.8 && feature.UniquePageCount < 5 && feature.Duration > 240 { // Less unique pages but longer time suggests aimlessness
		return ArchetypeFrustrated
	}

	// Rule for "Targeted Visitor"
	if feature.Duration <= 60 && feature.PageCount <= 3 {
		return ArchetypeTargeted
	}

	// Rule for "Engaged Browser"
	// High duration and low looping
	if feature.Duration > 600 && feature.LoopScore < 1.5 { // Long session, efficient navigation
		return ArchetypeEngaged
	}
	// Many unique pages visited, also efficient
	if feature.UniquePageCount >= 7 && feature.LoopScore < 1.8 {
		return ArchetypeEngaged
	}
	// Default catch-all
	return ArchetypeDefault
}

// GetArchetypes is the main function to generate behavioral archetypes from session data.
func GetArchetypes(site string, from, to time.Time) ([]structs.Archetype, error) {
	sessionFeatures, err := GetSessionFeatures(site, from, to)
	if err != nil {
		return nil, err
	}

	if len(sessionFeatures) == 0 {
		return []structs.Archetype{}, nil
	}

	archetypeCounts := make(map[string]int)
	archetypeExamples := make(map[string]string)
	for _, feature := range sessionFeatures {
		archetype := classifySession(feature)
		archetypeCounts[archetype]++
		// Store the first session ID we find as an example
		if _, ok := archetypeExamples[archetype]; !ok {
			archetypeExamples[archetype] = feature.SessionID
		}
	}

	// Define characteristics for each archetype
	characteristicsMap := map[string][]structs.ArchetypeCharacteristic{
		ArchetypeTargeted: {
			{Name: "Munkamenet hossza", Value: "Rövid (< 1 perc)"},
			{Name: "Viselkedés", Value: "Kevés (1-3) oldalt néz meg"},
			{Name: "Feltételezett Cél", Value: "Gyors információszerzés"},
		},
		ArchetypeEngaged: {
			{Name: "Munkamenet hossza", Value: "Hosszú (> 10 perc)"},
			{Name: "Viselkedés", Value: "Sok oldalt néz meg, lineárisan halad"},
			{Name: "Feltételezett Cél", Value: "Mélyreható kutatás, böngészés"},
		},
		ArchetypeFrustrated: {
			{Name: "Munkamenet hossza", Value: "Változó"},
			{Name: "Viselkedés", Value: "Sokszor visszalép, körbe-körbe jár"},
			{Name: "Feltételezett Cél", Value: "Nem találja, amit keres"},
		},
		ArchetypeDefault: {
			{Name: "Munkamenet hossza", Value: "Átlagos"},
			{Name: "Viselkedés", Value: "Általános böngészési minták"},
			{Name: "Feltételezett Cél", Value: "Vegyes"},
		},
	}

	var response []structs.Archetype
	totalSessions := float64(len(sessionFeatures))
	for name, count := range archetypeCounts {
		percentage := (float64(count) / totalSessions) * 100
		response = append(response, structs.Archetype{
			Name:             name,
			Percentage:       math.Round(percentage*10) / 10,
			Characteristics:  characteristicsMap[name],
			ExampleSessionID: archetypeExamples[name],
		})
	}

	// Sort by percentage descending
	sort.Slice(response, func(i, j int) bool {
		return response[i].Percentage > response[j].Percentage
	})

	return response, nil
}
