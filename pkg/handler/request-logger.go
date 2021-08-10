package handler

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func RequestLogger(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, h)
}
