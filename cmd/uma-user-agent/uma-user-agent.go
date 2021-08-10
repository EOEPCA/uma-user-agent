package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/EOEPCA/uma-user-agent/pkg/handler"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info(filepath.Base(os.Args[0]), " STARTING")

	router := mux.NewRouter()

	// Register middlewares
	router.Use(handler.RequestLogger)

	// Register request handler
	router.PathPrefix("").HandlerFunc(handler.NginxAuthRequestHandler)

	// Start listening
	port := config.Config.Port
	log.Info("Begin listening on port ", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), router))
}
