package analysis

import (
	"math"
	"statistics/structs"
	"time"
)

type alertThresholds struct {
	FrustratedShare float64
	BounceRate      float64
	EngagementDrop  float64
}

var defaultAlertThresholds = alertThresholds{
	FrustratedShare: 35,
	BounceRate:      60,
	EngagementDrop:  15,
}

// GenerateAlerts combines archetype mix, engagement scores, and funnel data to produce alert hints.
func GenerateAlerts(archetypes []structs.Archetype, currentEngagement float64, previousEngagement float64, bounceRate float64) []structs.ArchetypeAlert {
	threshold := defaultAlertThresholds
	var alerts []structs.ArchetypeAlert

	frustrated := findArchetype(archetypes, "Frusztrált vagy Céltalan")
	if frustrated != nil && frustrated.Percentage >= threshold.FrustratedShare {
		alerts = append(alerts, structs.ArchetypeAlert{
			Level:     "warning",
			Message:   "A frusztrált felhasználók aránya meghaladja a 35%-ot – ellenőrizd a hibákat vagy UX akadályokat.",
			Metric:    "Frusztrált archetípus aránya",
			Value:     frustrated.Percentage,
			Threshold: threshold.FrustratedShare,
		})
	}

	if bounceRate >= threshold.BounceRate {
		alerts = append(alerts, structs.ArchetypeAlert{
			Level:     "critical",
			Message:   "Kiugróan magas visszafordulási arány – a landing oldal nem konvertál.",
			Metric:    "Bounce rate",
			Value:     bounceRate,
			Threshold: threshold.BounceRate,
		})
	}

	if previousEngagement > 0 {
		drop := percentageDrop(previousEngagement, currentEngagement)
		if drop >= threshold.EngagementDrop {
			alerts = append(alerts, structs.ArchetypeAlert{
				Level:     "warning",
				Message:   "Az elköteleződési pontszám meredeken csökkent – vizsgáld meg az új tartalmakat vagy hibákat.",
				Metric:    "Engagement score",
				Value:     drop,
				Threshold: threshold.EngagementDrop,
			})
		}
	}

	return alerts
}

func findArchetype(archetypes []structs.Archetype, name string) *structs.Archetype {
	for _, a := range archetypes {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

func percentageDrop(previous, current float64) float64 {
	if previous == 0 {
		return 0
	}
	return math.Max(0, (previous-current)/previous*100)
}

// GetAlertWindow determines default time ranges used by analytics panels.
func GetAlertWindow(from, to time.Time) (time.Time, time.Time) {
	if from.IsZero() || to.IsZero() {
		to = time.Now()
		from = to.AddDate(0, 0, -7)
	}
	return from, to
}
