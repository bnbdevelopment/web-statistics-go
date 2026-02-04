package analysis

import (
	"fmt"
	"math"
	"sort"
	"statistics/database"
	"statistics/structs"
	"time"
)

const (
	ArchetypeTargeted     = "Célirányos Látogató"
	ArchetypeEngaged      = "Elmélyült Böngésző"
	ArchetypeFrustrated   = "Frusztrált vagy Céltalan"
	ArchetypeDefault      = "Általános Látogató"
)

// sessionFeature holds the calculated behavioral metrics for a single session.
type sessionFeature struct {
	SessionID       string
	Duration        float64 // in seconds
	PageCount       int
	UniquePageCount int
	LoopScore       float64 // Ratio of total pages to unique pages
}

// getSessionFeatures queries and calculates behavioral features for all sessions in a given timeframe.
func getSessionFeatures(site string, from, to time.Time) ([]sessionFeature, error) {
	type rawSessionData struct {
		SessionID       string
		StartTime       time.Time
		EndTime         time.Time
		PageCount       int
		UniquePageCount int
	}

	var rawData []rawSessionData

	dbQuery := database.Session.
		Model(&structs.WebMetric{}).
		Select(`
            session_id,
            MIN(timestamp) as start_time,
            MAX(timestamp) as end_time,
            COUNT(*) as page_count,
            COUNT(DISTINCT page) as unique_page_count
        `).
		Where("timestamp BETWEEN ? AND ?", from, to)

	if site != "" {
		dbQuery = dbQuery.Where("site = ?", site)
	}

	err := dbQuery.Group("session_id").Find(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query session data: %w", err)
	}

	var features []sessionFeature
	for _, data := range rawData {
		duration := data.EndTime.Sub(data.StartTime).Seconds()
		var loopScore float64
		if data.UniquePageCount > 0 {
			loopScore = float64(data.PageCount) / float64(data.UniquePageCount)
		}

		features = append(features, sessionFeature{
			SessionID:       data.SessionID,
			Duration:        duration,
			PageCount:       data.PageCount,
			UniquePageCount: data.UniquePageCount,
			LoopScore:       loopScore,
		})
	}
	return features, nil
}

// classifySession applies a set of rules to categorize a session into an archetype.
func classifySession(feature sessionFeature) string {
	// Rule for "Frustrated or Aimless"
	if feature.LoopScore >= 2.5 && feature.PageCount > 5 {
		return ArchetypeFrustrated
	}
	if feature.Duration > 120 && feature.PageCount > 10 && feature.LoopScore >= 1.8 {
		return ArchetypeFrustrated
	}

	// Rule for "Targeted Visitor"
	if feature.Duration <= 60 && feature.PageCount <= 3 {
		return ArchetypeTargeted
	}

	// Rule for "Engaged Browser"
	if feature.Duration > 600 && feature.LoopScore < 1.5 {
		return ArchetypeEngaged
	}
	if feature.UniquePageCount > 7 && feature.LoopScore < 1.8 {
		return ArchetypeEngaged
	}

	// Default catch-all
	return ArchetypeDefault
}

// GetArchetypes is the main function to generate behavioral archetypes from session data.
func GetArchetypes(site string, from, to time.Time) ([]structs.Archetype, error) {
	sessionFeatures, err := getSessionFeatures(site, from, to)
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
			Percentage:       math.Round(percentage*10)/10,
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
