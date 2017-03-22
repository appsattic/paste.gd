package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gomiddleware/logger"
	"github.com/gomiddleware/logit"
	"github.com/gomiddleware/mux"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// setup the logger
	lgr := logit.New(os.Stdout, "paste")

	// setup
	apex := os.Getenv("PASTE_APEX")
	baseUrl := os.Getenv("PASTE_BASE_URL")
	port := os.Getenv("PASTE_PORT")
	if port == "" {
		log.Fatal("Specify a port to listen on in the environment variable 'PASTE_PORT'")
	}

	// load up all templates
	tmpl, err := template.New("").ParseGlob("./templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	// the mux
	m := mux.New()

	m.Use("/", logger.NewLogger(lgr))

	// do some static routes before doing logging
	m.All("/s", fileServer("static"))
	m.Get("/favicon.ico", serveFile("./static/favicon.ico"))
	m.Get("/robots.txt", serveFile("./static/robots.txt"))

	m.Get("/sitemap.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, baseUrl+"/\n")
	})

	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Apex    string
			BaseUrl string
		}{
			apex,
			baseUrl,
		}
		render(w, tmpl, "index.html", data)
	})

	// finally, check all routing was added correctly
	check(m.Err)

	// server
	fmt.Printf("Starting server, listening on port %s\n", port)
	errServer := http.ListenAndServe(":"+port, m)
	check(errServer)
}
