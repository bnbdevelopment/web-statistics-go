package config

import (
	"testing"
)

func TestGetFunnelSteps(t *testing.T) {
	steps := GetFunnelSteps("unknown-site.hu")
	if len(steps) == 0 {
		t.Fatalf("Expected default funnel steps, got empty slice")
	}

	for i, step := range steps {
		if step.Label == "" {
			t.Errorf("Step %d has empty label", i)
		}
		if len(step.Keywords) == 0 {
			t.Errorf("Step %d (%s) has no keywords", i, step.Label)
		}
	}
}
