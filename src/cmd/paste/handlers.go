package main

// https://gist.github.com/chilts/db1adfaddaae871b161d7eadab6b1278

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

func serveFile(filename string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}

func fileServer(dirname string) http.Handler {
	return http.FileServer(http.Dir(dirname))
}

func redirect(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, http.StatusFound)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func internalServerError(w http.ResponseWriter, err error) {
	log.Printf("Err: %s\n", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func render(w http.ResponseWriter, tmpl *template.Template, tmplName string, data interface{}) {
	buf := &bytes.Buffer{}
	err := tmpl.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}
