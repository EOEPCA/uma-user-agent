package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, h)
}

func main() {
	log.Info(filepath.Base(os.Args[0]), " STARTING")

	router := mux.NewRouter()

	// Register middlewares
	router.Use(loggingMiddleware)

	router.PathPrefix("").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this is the uma-user-agent")
	})

	http.ListenAndServe(":8080", router)
}
