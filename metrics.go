package main

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

const chosenInlineResultCounter = "inline_result_was_chosen"

type handlerName string

var handlerCounters = make(map[handlerName]prometheus.Counter)

func init() {
	registerCounter(chosenInlineResultCounter, "")
}

func registerMessageHandlerCounters(handlers ...base.MessageHandler) {
	for _, h := range handlers {
		registerCounter(resolveHandlerName(h), "used_message_handler_")
	}
}

func incMessageHandlerCounter(handler base.MessageHandler) {
	inc(resolveHandlerName(handler))
}

func registerInlineHandlerCounters(handlers ...base.InlineHandler) {
	for _, h := range handlers {
		registerCounter(resolveHandlerName(h), "used_inline_handler_")
	}
}

func incInlineHandlerCounter(handler base.InlineHandler) {
	inc(resolveHandlerName(handler))
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

func inc(name string) {
	counter, ok := handlerCounters[handlerName(name)]
	if ok {
		counter.Inc()
	} else {
		log.WithField(logconst.FieldFunc, "inc").
			Warning("Counter " + name + " is missing!")
	}
}

func addHttpHandlerForMetrics() {
	http.Handle("/metrics", promhttp.Handler())
}
