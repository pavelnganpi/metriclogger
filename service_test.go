package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanUpExpiredMetrics(t *testing.T) {
	key, metricAgg := getTestMetricAgg()

	s := MetricsService{
		Store: map[string]*MetricsAgg{key: metricAgg},
	}

	assert.Equal(t, 3, len(metricAgg.metrics))
	s.cleanUpExpiredMetrics(key)

	assert.Equal(t, 2, len(metricAgg.metrics))
	assert.Equal(t, int64(3), metricAgg.count)
	assert.True(t, time.Since(metricAgg.startTime) < time.Minute) // assert new start time is very recent
}

func TestGetMetricSum(t *testing.T) {
	key, metricAgg := getTestMetricAgg()
	s := MetricsService{
		Store: map[string]*MetricsAgg{key: metricAgg},
	}

	assert.Equal(t, int64(8), s.GetMetricSum(key))
}

func TestProcessMetrics(t *testing.T) {
	metric1 := Metric{
		eventAt: time.Now(),
		key:     "paigeview",
		value:   5,
	}
	metric2 := Metric{
		eventAt: time.Now(),
		key:     "logins",
		value:   2,
	}
	metric3 := Metric{
		eventAt: time.Now(),
		key:     "logins",
		value:   1,
	}

	s := NewMetricsService()
	s.mCh <- metric1
	s.mCh <- metric2
	s.mCh <- metric3

	time.Sleep(50 *time.Millisecond) // give room for goroutine to process metrics

	assert.Equal(t, 0, len(s.mCh))
	assert.Equal(t, int64(5), s.Store["paigeview"].count)
	assert.Equal(t, int64(3), s.Store["logins"].count)

	assert.Equal(t, []Metric{metric1}, s.Store["paigeview"].metrics)
	assert.Equal(t, []Metric{metric2, metric3}, s.Store["logins"].metrics)

	s.Shutdown()
}

func TestCreateTicker(t *testing.T) {
	cleanUpInterval = 25 * time.Millisecond
	defer func() {
		cleanUpInterval = time.Hour
	}()

	key := "paigeview"
	metric1 := Metric{
		eventAt: time.Now().Add(-61 * time.Minute), // expired
		key:     key,
		value:   5,
	}
	metricAgg := MetricsAgg{
		metrics:   []Metric{metric1},
		count:     5,
	}
	s := MetricsService{
		Store: map[string]*MetricsAgg{key: &metricAgg},
	}

	s.createTicker(key)

	time.Sleep(50 *time.Millisecond) // give room for goroutine to execute

	assert.Equal(t, 0, len(s.mCh))
	assert.Equal(t, int64(0), s.Store[key].count)
	assert.Equal(t, 0, len(s.Store[key].metrics))
	assert.True(t, time.Since(s.Store[key].startTime) < time.Minute) // assert new start time is very recent
}

func getTestMetricAgg() (string, *MetricsAgg) {
	key := "pageview"
	metric1 := Metric{
		eventAt: time.Now().Add(-61 * time.Minute), // should expire and be removed on clean up
		key:     key,
		value:   5,
	}
	metric2 := Metric{
		eventAt: time.Now().Add(-52 * time.Minute),
		key:     key,
		value:   2,
	}
	metric3 := Metric{
		eventAt: time.Now().Add(-6 * time.Second),
		key:     key,
		value:   1,
	}
	metricAgg := MetricsAgg{
		metrics:   []Metric{metric1, metric2, metric3},
		count:     8,
	}
	return key, &metricAgg
}
