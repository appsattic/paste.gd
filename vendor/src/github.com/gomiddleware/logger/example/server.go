package main

import (
	"net/http"
	"os"

	"github.com/gomiddleware/logger"
	"github.com/gomiddleware/logit"
)

func handler(lgr *logit.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lgr := logger.LogFromRequest(r)
		lgr.Print("handler-start")

		w.Write([]byte(r.URL.Path))
	}
}

func main() {
	// create the logger middleware
	lgr := logit.New(os.Stdout, "main")
	log := logger.NewLogger(lgr)

	// make the http.Hander and wrap it with the log middleware
	handle := http.HandlerFunc(handler(lgr))
	http.Handle("/", log(handle))

	http.ListenAndServe(":8080", nil)
}
