package analysis

import (
	"statistics/structs"
	"testing"
)

func TestGenerateAlerts(t *testing.T) {
	// Scenario 1: All healthy - no alerts expected
	healthyArchetypes := []structs.Archetype{
		{Name: "Célirányos Látogató", Percentage: 50},
		{Name: "Elmélyült Böngésző", Percentage: 35},
		{Name: "Frusztrált vagy Céltalan", Percentage: 15},
	}
	alerts := GenerateAlerts(healthyArchetypes, 65.0, 60.0, 30.0)
	if len(alerts) != 0 {
		t.Errorf("Expected 0 alerts for healthy metrics, got %d", len(alerts))
	}

	// Scenario 2: High bounce rate (> 60%)
	alertsBounce := GenerateAlerts(healthyArchetypes, 65.0, 60.0, 75.0)
	if len(alertsBounce) != 1 {
		t.Fatalf("Expected 1 alert for high bounce rate, got %d", len(alertsBounce))
	}
	if alertsBounce[0].Level != "critical" || alertsBounce[0].Metric != "Bounce rate" {
		t.Errorf("Unexpected alert properties: %+v", alertsBounce[0])
	}

	// Scenario 3: High frustrated visitors (> 35%)
	frustratedArchetypes := []structs.Archetype{
		{Name: "Frusztrált vagy Céltalan", Percentage: 45},
	}
	alertsFrustrated := GenerateAlerts(frustratedArchetypes, 65.0, 60.0, 30.0)
	if len(alertsFrustrated) != 1 {
		t.Fatalf("Expected 1 alert for frustrated visitors, got %d", len(alertsFrustrated))
	}
	if alertsFrustrated[0].Level != "warning" {
		t.Errorf("Expected warning alert, got %s", alertsFrustrated[0].Level)
	}

	// Scenario 4: Steep engagement score drop (> 15%)
	alertsDrop := GenerateAlerts(healthyArchetypes, 40.0, 80.0, 30.0)
	if len(alertsDrop) != 1 {
		t.Fatalf("Expected 1 alert for engagement score drop, got %d", len(alertsDrop))
	}
	if alertsDrop[0].Metric != "Engagement score" {
		t.Errorf("Expected 'Engagement score' alert, got %s", alertsDrop[0].Metric)
	}
}
