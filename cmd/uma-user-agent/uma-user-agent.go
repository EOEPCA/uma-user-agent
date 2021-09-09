package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/EOEPCA/uma-user-agent/pkg/handler"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func configSummry() {
	go func() {
		for {
			fmt.Println("client-id =", config.GetClientId())
			time.Sleep(time.Second * 2)
		}
	}()
}

func main() {
	logrus.Info(filepath.Base(os.Args[0]), " STARTING")

	configSummry()

	router := mux.NewRouter()

	// Register middlewares
	router.Use(handler.RequestLogger)

	// Register request handler for status
	handler.NewStatusRouter(router.PathPrefix("/status").Subrouter())

	// Register request handler for auth_request
	router.PathPrefix("").HandlerFunc(handler.NginxAuthRequestHandler)

	// Start listening
	port := config.GetPort()
	logrus.Info("Begin listening on port ", port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), router))
}
