package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/EOEPCA/uma-user-agent/pkg/handler"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info(filepath.Base(os.Args[0]), " STARTING")

	router := mux.NewRouter()

	// Register middlewares
	// COMMENTED OUT, since the usual Request Logger is less useful in this case
	// because the 'target' URL is passed in the http headers
	// router.Use(handler.RequestLogger)

	// Register request handler for status
	handler.NewStatusRouter(router.PathPrefix("/status").Subrouter())

	// Register request handler for auth_request
	router.PathPrefix("").HandlerFunc(handler.NginxAuthRequestHandler)

	// Start listening
	port := config.GetPort()
	logrus.Info("Begin listening on port ", port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), router))
}
