package handlers

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type fakeRowsAffected int64

func (n fakeRowsAffected) RowsAffected() int64 {
	return int64(n)
}

// assertIncremented checks that the action increments the counter by one.
func assertIncremented(t *testing.T, counter prometheus.Counter, action func()) {
	before := testutil.ToFloat64(counter)
	action()
	assert.Equal(t, before+1, testutil.ToFloat64(counter))
}

func TestCountSaveOutcome(t *testing.T) {
	outcome := func(o string) prometheus.Counter {
		return saveOutcomeCounter.WithLabelValues(metricModeCommand, o)
	}
	assertIncremented(t, outcome(metricOutcomeSuccess), func() {
		countSaveOutcome(metricModeCommand, fakeRowsAffected(1), nil)
	})
	assertIncremented(t, outcome(metricOutcomeDuplicate), func() {
		countSaveOutcome(metricModeCommand, nil, &pgconn.PgError{Code: "23505"})
	})
	assertIncremented(t, outcome(metricOutcomeFailure), func() {
		countSaveOutcome(metricModeCommand, nil, errors.New("connection refused"))
	})
	assertIncremented(t, outcome(metricOutcomeFailure), func() {
		countSaveOutcome(metricModeCommand, fakeRowsAffected(0), nil)
	})
}

func TestCountShowCacheGet(t *testing.T) {
	result := func(r string) prometheus.Counter {
		return showCacheCounter.WithLabelValues(metricCacheGet, r)
	}
	assertIncremented(t, result(metricCacheHit), func() { countShowCacheGet(nil) })
	assertIncremented(t, result(metricCacheMiss), func() { countShowCacheGet(ErrShowCacheMiss) })
	assertIncremented(t, result(metricCacheError), func() { countShowCacheGet(errors.New("redis is down")) })
}

func TestCountSearch(t *testing.T) {
	assertIncremented(t, searchCounter.WithLabelValues(metricModeShow, metricResultFound), func() { countSearch(metricModeShow, 2) })
	assertIncremented(t, searchCounter.WithLabelValues(metricModeShow, metricResultEmpty), func() { countSearch(metricModeShow, 0) })
}

func TestCountersAreInitialized(t *testing.T) {
	// 3 modes × 2 results, 2 modes × 3 outcomes, 4 outcomes, 2 modes, 3 + 2 results of the cache
	assert.GreaterOrEqual(t, testutil.CollectAndCount(searchCounter), 6)
	assert.GreaterOrEqual(t, testutil.CollectAndCount(saveOutcomeCounter), 6)
	assert.GreaterOrEqual(t, testutil.CollectAndCount(installOutcomeCounter), 4)
	assert.GreaterOrEqual(t, testutil.CollectAndCount(carouselCounter), 2)
	assert.GreaterOrEqual(t, testutil.CollectAndCount(showCacheCounter), 5)
}
