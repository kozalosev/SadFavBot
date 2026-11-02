package handlers

import (
	"errors"

	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// label values of the metrics below
const (
	metricModeCommand = "command"
	metricModeInline  = "inline"
	metricModeGuest   = "guest"
	metricModeShow    = "show"

	metricOutcomeSuccess   = "success"
	metricOutcomeDuplicate = "duplicate"
	metricOutcomeFailure   = "failure"
	metricOutcomeNoRows    = "no_rows"
	metricOutcomeLinkToFav = "link_to_existing_fav"

	metricResultFound = "found"
	metricResultEmpty = "empty"

	metricCacheGet   = "get"
	metricCachePut   = "put"
	metricCacheHit   = "hit"
	metricCacheMiss  = "miss"
	metricCacheOK    = "ok"
	metricCacheError = "error"
)

var (
	carouselCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "carousel_shown_total",
		Help: "Count of messages sent with the buttons to switch between favs, by mode (show/guest). Read it against used_callback_handler_* to see how much the buttons are used",
	}, []string{"mode"})
	searchCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "search_total",
		Help: "Count of searches of favs by alias, by mode (inline/guest/show) and result (found/empty). Only the first page of an inline query is counted",
	}, []string{"mode", "result"})
	saveOutcomeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "save_outcome_total",
		Help: "Count of attempts to save a fav, by mode (command/guest) and outcome (success/duplicate/failure)",
	}, []string{"mode", "outcome"})
	installOutcomeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "install_outcome_total",
		Help: "Count of attempts to install a package, by outcome (success/no_rows/link_to_existing_fav/failure)",
	}, []string{"outcome"})
	showCacheCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "show_cache_total",
		Help: "Count of operations with the cache of the carousel buttons, by operation (get/put) and result (hit/miss/error for get, ok/error for put). A lot of misses means the buttons expire too early: increase SHOW_CACHE_TTL_SECS",
	}, []string{"operation", "result"})
)

func countSearch(mode string, found int) {
	result := metricResultEmpty
	if found > 0 {
		result = metricResultFound
	}
	searchCounter.WithLabelValues(mode, result).Inc()
}

// countSaveOutcome counts the result of [repo.FavService.Save] the same way the save handlers interpret it.
func countSaveOutcome(mode string, res repo.RowsAffectedAware, err error) {
	var outcome string
	switch {
	case err != nil && storage.DuplicateConstraintViolation(err):
		outcome = metricOutcomeDuplicate
	case err != nil, res.RowsAffected() == 0:
		outcome = metricOutcomeFailure
	default:
		outcome = metricOutcomeSuccess
	}
	saveOutcomeCounter.WithLabelValues(mode, outcome).Inc()
}

func countShowCacheGet(err error) {
	result := metricCacheHit
	if errors.Is(err, ErrShowCacheMiss) {
		result = metricCacheMiss
	} else if err != nil {
		result = metricCacheError
	}
	showCacheCounter.WithLabelValues(metricCacheGet, result).Inc()
}

func countShowCachePut(err error) {
	result := metricCacheOK
	if err != nil {
		result = metricCacheError
	}
	showCacheCounter.WithLabelValues(metricCachePut, result).Inc()
}
