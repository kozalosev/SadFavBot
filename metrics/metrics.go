// Package metrics provides metrics for Prometheus.
package metrics

import (
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/logconst"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

// ChosenInlineResultCounter is a counter which is registered automatically for ChosenInlineResult.
// https://core.telegram.org/bots/inline#collecting-feedback
const ChosenInlineResultCounter = "inline_result_was_chosen"

type handlerName string

var handlerCounters = make(map[handlerName]prometheus.Counter)

func init() {
	registerCounter(ChosenInlineResultCounter, "")
}

// RegisterMessageHandlerCounters following the scheme: "used_message_handler_%s".
func RegisterMessageHandlerCounters(handlers ...base.MessageHandler) {
	for _, h := range handlers {
		registerCounter(resolveHandlerName(h), "used_message_handler_")
	}
}

// IncMessageHandlerCounter increments the counter registered for a message handler.
func IncMessageHandlerCounter(handler base.MessageHandler) {
	Inc(resolveHandlerName(handler))
}

// RegisterInlineHandlerCounters following the scheme: "used_inline_handler_%s".
func RegisterInlineHandlerCounters(handlers ...base.InlineHandler) {
	for _, h := range handlers {
		registerCounter(resolveHandlerName(h), "used_inline_handler_")
	}
}

// IncInlineHandlerCounter increments the counter registered for an inline query handler.
func IncInlineHandlerCounter(handler base.InlineHandler) {
	Inc(resolveHandlerName(handler))
}

// Inc increments the counter registered as 'name'.
func Inc(name string) {
	counter, ok := handlerCounters[handlerName(name)]
	if ok {
		counter.Inc()
	} else {
		log.WithField(logconst.FieldFunc, "Inc").
			Warning("Counter " + name + " is missing!")
	}
}

// AddHttpHandlerForMetrics adds a global route /metrics to the server by using [http.Handle].
func AddHttpHandlerForMetrics() {
	http.Handle("/metrics", promhttp.Handler())
}

func resolveHandlerName(handler any) string {
	t := reflect.TypeOf(handler)
	if t.Kind() == reflect.Pointer {
		return reflect.Indirect(reflect.ValueOf(handler)).Type().Name()
	} else {
		return t.Name()
	}
}

func registerCounter(name, metricPrefix string) {
	handlerCounters[handlerName(name)] = promauto.NewCounter(prometheus.CounterOpts{
		Name: metricPrefix + name,
		Help: "Usage counter",
	})
}
