// Package server provides functions to set a webhook and start a server to process incoming requests.
package server

import (
	"context"
	"github.com/kozalosev/SadFavBot/logconst"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Start a server listening on the specified port.
func Start(port string) *http.Server {
	srv := &http.Server{Addr: ":" + port}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.WithField(logconst.FieldFunc, "startServer").
				WithField(logconst.FieldCalledObject, "Server").
				WithField(logconst.FieldCalledMethod, "ListenAndServe").
				Fatal(err)
		}
	}()
	return srv
}

// StopListeningForIncomingRequests stops the server.
func StopListeningForIncomingRequests(srv *http.Server) {
	ctx, c := context.WithTimeout(context.Background(), time.Minute)
	defer c()
	if err := srv.Shutdown(ctx); err != nil {
		log.WithField(logconst.FieldFunc, "stopListeningForIncomingRequests").
			WithField(logconst.FieldCalledObject, "Server").
			WithField(logconst.FieldCalledMethod, "Shutdown").
			Error(err)
	}
}
