package database

import (
	"log"
	"statistics/structs"
	"sync"
	"sync/atomic"
	"time"
)

// IngestionQueue handles asynchronous, high-throughput batched ingestion of web metrics.
type IngestionQueue struct {
	queue           chan structs.WebMetric
	batchSize       int
	flushInterval   time.Duration
	wg              sync.WaitGroup
	quit            chan struct{}
	totalIngested   uint64
	fallbackInserts uint64
	batchErrors     uint64
}

var GlobalIngestionQueue *IngestionQueue

// InitIngestionQueue initializes and starts the background batch flusher.
func InitIngestionQueue(capacity int, batchSize int, flushInterval time.Duration) *IngestionQueue {
	if capacity <= 0 {
		capacity = 10000
	}
	if batchSize <= 0 {
		batchSize = 500
	}
	if flushInterval <= 0 {
		flushInterval = 500 * time.Millisecond
	}

	q := &IngestionQueue{
		queue:         make(chan structs.WebMetric, capacity),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		quit:          make(chan struct{}),
	}

	GlobalIngestionQueue = q

	q.wg.Add(1)
	go q.worker()

	log.Printf("Ingestion queue initialized (capacity: %d, batch: %d, flush: %v)", capacity, batchSize, flushInterval)
	return q
}

// Enqueue queues a WebMetric record for batched asynchronous persistence.
func (q *IngestionQueue) Enqueue(metric structs.WebMetric) bool {
	atomic.AddUint64(&q.totalIngested, 1)
	select {
	case q.queue <- metric:
		return true
	default:
		// Queue full: fallback to direct insert to prevent metric loss under burst load
		atomic.AddUint64(&q.fallbackInserts, 1)
		log.Println("Ingestion queue full, performing synchronous fallback insert")
		if Session != nil {
			if err := Session.Create(&metric).Error; err != nil {
				log.Printf("Fallback insert failed: %v", err)
				return false
			}
		}
		return true
	}
}

// worker collects records and flushes them periodically or when batch size is reached.
func (q *IngestionQueue) worker() {
	defer q.wg.Done()

	ticker := time.NewTicker(q.flushInterval)
	defer ticker.Stop()

	batch := make([]structs.WebMetric, 0, q.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if Session != nil {
			if err := Session.CreateInBatches(batch, len(batch)).Error; err != nil {
				atomic.AddUint64(&q.batchErrors, 1)
				log.Printf("Batch insertion error (%d items): %v", len(batch), err)
			}
		}
		batch = make([]structs.WebMetric, 0, q.batchSize)
	}

	for {
		select {
		case item, ok := <-q.queue:
			if !ok {
				flush()
				return
			}
			batch = append(batch, item)
			if len(batch) >= q.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-q.quit:
			// Drain remaining items from queue
			for {
				select {
				case item := <-q.queue:
					batch = append(batch, item)
					if len(batch) >= q.batchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

// Stop flushes all pending records and shuts down the background worker.
func (q *IngestionQueue) Stop() {
	if q == nil {
		return
	}
	close(q.quit)
	q.wg.Wait()
}

// GetStats returns current metrics for queue monitoring.
func (q *IngestionQueue) GetStats() (queueLen int, queueCap int, totalIngested uint64, fallbackInserts uint64, batchErrors uint64) {
	if q == nil {
		return 0, 0, 0, 0, 0
	}
	return len(q.queue), cap(q.queue), atomic.LoadUint64(&q.totalIngested), atomic.LoadUint64(&q.fallbackInserts), atomic.LoadUint64(&q.batchErrors)
}
