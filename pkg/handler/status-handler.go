package handler

import (
	"fmt"
	"net/http"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/gorilla/mux"
)

// NewStatusRouter registers the handlers to report service status for probes.
func NewStatusRouter(router *mux.Router) *mux.Router {

	// Readiness
	router.PathPrefix("/ready").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Config.IsReady() {
			fmt.Fprintln(w, "READY")
		} else {
			w.WriteHeader(http.StatusTooEarly)
			fmt.Fprintln(w, "NOT READY")
		}
	})

	// Liveness
	router.PathPrefix("/alive").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ALIVE")
	})

	return router
}
