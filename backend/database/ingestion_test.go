package database

import (
	"statistics/structs"
	"testing"
	"time"
)

func TestIngestionQueueOperations(t *testing.T) {
	// Initialize queue with capacity 100, batch size 20, long flush interval so items stay buffered
	queue := InitIngestionQueue(100, 20, 10*time.Second)
	defer queue.Stop()

	// Initial stats
	qLen, qCap, totalIngested, fallbacks, batchErrors := queue.GetStats()
	if qCap != 100 {
		t.Errorf("Expected capacity 100, got %d", qCap)
	}
	if totalIngested != 0 || fallbacks != 0 || batchErrors != 0 {
		t.Errorf("Expected zero initial metrics, got total=%d, fallbacks=%d, errors=%d", totalIngested, fallbacks, batchErrors)
	}

	// Enqueue items
	for i := 0; i < 5; i++ {
		metric := structs.WebMetric{
			SessionId: "test-session",
			Site:      "example.com",
			Page:      "/test",
			Timestamp: time.Now(),
		}
		ok := queue.Enqueue(metric)
		if !ok {
			t.Errorf("Failed to enqueue metric %d", i)
		}
	}

	// Verify updated metrics
	qLen, _, totalIngested, _, _ = queue.GetStats()
	if totalIngested != 5 {
		t.Errorf("Expected 5 total ingested, got %d", totalIngested)
	}
	if qLen < 0 || qLen > 5 {
		t.Errorf("Unexpected queue length %d", qLen)
	}
}

func TestIngestionQueueFullFallback(t *testing.T) {
	// Initialize small queue with capacity 2 and batch size 10 (never flushes automatically during test)
	queue := InitIngestionQueue(2, 10, 10*time.Second)
	defer queue.Stop()

	// Fill queue
	queue.Enqueue(structs.WebMetric{SessionId: "s1"})
	queue.Enqueue(structs.WebMetric{SessionId: "s2"})

	// This 3rd item should trigger fallback since queue buffer is full (capacity 2)
	queue.Enqueue(structs.WebMetric{SessionId: "s3"})

	_, _, totalIngested, fallbacks, _ := queue.GetStats()
	if totalIngested != 3 {
		t.Errorf("Expected 3 total ingested, got %d", totalIngested)
	}
	if fallbacks < 1 {
		t.Errorf("Expected at least 1 fallback insert, got %d", fallbacks)
	}
}
