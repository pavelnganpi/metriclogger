package main

import (
	"sync"
	"time"
)

var cleanUpInterval = time.Hour

// MetricsAgg is a struct that tracks and aggregates metrics for a specific metric key
// Keeps log of metrics for the last hour
type MetricsAgg struct {
	metrics   []Metric
	count     int64
	startTime time.Time
	sync.Mutex
}

// Metric represents a metric event been logged
type Metric struct {
	eventAt time.Time // time metric was received
	key     string    // metric name, id
	value   int64     // value representation of a metric
}

// MetricsService tracks and stores a stream of metrics, keeps track of metrics within the last hour
type MetricsService struct {
	Store   map[string]*MetricsAgg // map to aggregate all metrics by key
	mCh     chan Metric            // channel to process metrics synchronously
	tickers []*time.Ticker         // stores list of tickers for each metric key, to gracefully close them when exiting
}

// NewMetricsService returns an instance of a MetricsService{}
func NewMetricsService() *MetricsService {
	s := &MetricsService{
		Store: make(map[string]*MetricsAgg),
		mCh:   make(chan Metric, 100),
	}

	go s.ProcessMetrics() // process metrics in an independent goroutine

	return s
}

// ProcessMetrics consumes Metric{} from the channel and stores them in a map
// Uses a goroutine for each metric key to manage/keep metrics that occurred in the lastest hour
func (s *MetricsService) ProcessMetrics() {
	for metric := range s.mCh {
		// create new metricsAgg if key is not present in map
		if _, ok := s.Store[metric.key]; !ok {
			now := time.Now()
			metricsAgg := initMetricsAgg()
			metricsAgg.startTime = now
			metricsAgg.count += metric.value
			metricsAgg.metrics = append(metricsAgg.metrics, metric)
			s.Store[metric.key] = metricsAgg

			// spin up a new goroutine to clean up metrics for this metric key
			s.createTicker(metric.key)

			// update metricsAgg
		} else {
			metricsAgg := s.Store[metric.key]
			metricsAgg.metrics = append(metricsAgg.metrics, metric)
			metricsAgg.count += metric.value

		}
	}
}

// createTicker creates a ticker instance to run periodically every hour
// to clean up expired metrics
func (s *MetricsService) createTicker(key string) {
	ticker := time.NewTicker(cleanUpInterval)

	go func() {
		for range ticker.C {
			s.cleanUpExpiredMetrics(key)
		}
	}()
	s.tickers = append(s.tickers, ticker)
}

// GetMetricSum returns the total metric value for the given key
func (s *MetricsService) GetMetricSum(key string) int64 {
	metricAgg := s.Store[key]
	if metricAgg != nil {
		metricAgg.Lock()
		defer metricAgg.Unlock()
		return metricAgg.count
	}
	return 0
}

// Shutdown closes all channels and tickers
func (s *MetricsService) Shutdown() {
	close(s.mCh)
	for _, ticker := range s.tickers {
		ticker.Stop()
	}
}

// cleanUpExpiredMetrics removes all metric events that are more than hour old,
// keeps metric count upto date for the latest hour
func (s *MetricsService) cleanUpExpiredMetrics(key string) {
	metricsAgg := s.Store[key]

	if metricsAgg != nil {
		metricsAgg.Lock()
		defer metricsAgg.Unlock()

		var expiredMetricsTotalValue int64
		var expiredMetricsCount int64
		// remove all expired metrics(> 1 hour old) if any
		for _, metric := range metricsAgg.metrics {
			if time.Since(metric.eventAt) > time.Hour {
				expiredMetricsTotalValue += metric.value
				expiredMetricsCount++
			} else {
				break
			}
		}
		metricsAgg.metrics = metricsAgg.metrics[expiredMetricsCount:] // remove all expired metrics
		metricsAgg.count -= expiredMetricsTotalValue                    // update current expiredMetricsTotalValue
		metricsAgg.startTime = time.Now()
	}
}

// initMetricsAgg initializes an empty instance of MetricsAgg{}
func initMetricsAgg() *MetricsAgg {
	return &MetricsAgg{
		metrics: make([]Metric, 0),
	}
}