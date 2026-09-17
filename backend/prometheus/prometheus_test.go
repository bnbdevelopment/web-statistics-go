package prometheus

import (
	"testing"

	promclient "github.com/prometheus/client_golang/prometheus"
)

func TestPrometheusMetricsRegistered(t *testing.T) {
	// Verify vectors are non-nil
	if visitorsBySite == nil || activeUsersBySite == nil || bounceRateBySite == nil {
		t.Fatal("Expected gauge vectors to be initialized")
	}

	// Set sample value to instantiate metric in gatherer
	visitorsBySite.With(promclient.Labels{"site": "test-site"}).Set(42)
	bounceRateBySite.With(promclient.Labels{"site": "test-site"}).Set(25.5)

	metricFamilies, err := promclient.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather Prometheus metrics: %v", err)
	}

	foundTraffic := false
	foundBounce := false
	foundQueue := false

	for _, mf := range metricFamilies {
		switch mf.GetName() {
		case "statistics_traffic":
			foundTraffic = true
		case "statistics_bounce_rate_percentage":
			foundBounce = true
		case "statistics_ingestion_queue_depth":
			foundQueue = true
		}
	}

	if !foundTraffic {
		t.Errorf("statistics_traffic was not gathered")
	}
	if !foundBounce {
		t.Errorf("statistics_bounce_rate_percentage was not gathered")
	}
	if !foundQueue {
		t.Errorf("statistics_ingestion_queue_depth was not gathered")
	}
}

func TestCollectAllMetricsNilSession(t *testing.T) {
	// Ensure that collectAllMetrics executes safely without panicking when database session is nil or empty
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("collectAllMetrics panicked with: %v", r)
		}
	}()

	collectAllMetrics()
}
