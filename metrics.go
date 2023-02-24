package main

import (
	"context"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"time"
)

const chosenInlineResultCounter = "inline_result_was_chosen"

type handlerName string

var handlerCounters = make(map[handlerName]prometheus.Counter)

func init() {
	registerCounter(chosenInlineResultCounter, "")
}

func registerMessageHandlerCounters(handlers ...base.MessageHandler) {
	for _, h := range handlers {
		registerCounter(reflect.TypeOf(h).Name(), "used_message_handler_")
	}
}

func incMessageHandlerCounter(handler base.MessageHandler) {
	inc(reflect.TypeOf(handler).Name())
}

func registerInlineHandlerCounters(handlers ...base.InlineHandler) {
	for _, h := range handlers {
		registerCounter(reflect.TypeOf(h).Name(), "used_inline_handler_")
	}
}

func incInlineHandlerCounter(handler base.InlineHandler) {
	inc(reflect.TypeOf(handler).Name())
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
		log.Warning("Counter " + name + " is missing!")
	}
}

func startMetricsServer(port string) *http.Server {
	srv := &http.Server{Addr: ":" + port}
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	return srv
}

func shutdownMetricsServer(metricsServer *http.Server) {
	ctx, c := context.WithTimeout(context.Background(), 1*time.Minute)
	defer c()
	if err := metricsServer.Shutdown(ctx); err != nil {
		log.Errorln(err)
	}
}
