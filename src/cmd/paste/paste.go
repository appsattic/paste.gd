package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chilts/rod"
	"github.com/gomiddleware/logger"
	"github.com/gomiddleware/logit"
	"github.com/gomiddleware/mux"
)

var pasteBucketNameStr = "paste"
var pasteBucketName = []byte(pasteBucketNameStr)

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
	dir := os.Getenv("PASTE_DIR")
	if dir == "" {
		log.Fatal("Specify a dir to write pastefiles to 'PASTE_DIR'")
	}

	// load up all templates
	tmpl, err := template.New("").ParseGlob("./templates/*.html")
	check(err)

	// open the datastore
	db, err := bolt.Open("paste.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	check(err)
	defer db.Close()

	// create the main buckets
	err = db.Update(func(tx *bolt.Tx) error {
		var err error

		_, err = tx.CreateBucketIfNotExists(pasteBucketName)
		if err != nil {
			return err
		}

		return nil
	})
	check(err)

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
			Paste   Paste
		}{
			apex,
			baseUrl,
			Paste{},
		}
		render(w, tmpl, "index.html", data)
	})

	m.Post("/paste", func(w http.ResponseWriter, r *http.Request) {
		// get the values of certain fields
		title := r.FormValue("title")
		text := r.FormValue("text")
		visibility := r.FormValue("visibility")
		if visibility != "unlisted" && visibility != "public" && visibility != "encrypted" {
			// since this comes from a form, then someone is messing with it - don't give them any leeway
			internalServerError(w, errors.New("visibility was not an allowed option"))
			return
		}

		// create the paste
		now := time.Now().UTC()
		paste := Paste{
			Id:         Id(6),
			Title:      title,
			Size:       len(text),
			Visibility: visibility,
			Created:    now,
			Updated:    now,
		}

		// save the text to a file
		filename := dir + "/" + paste.Id
		err = ioutil.WriteFile(filename, []byte(text), 0755)
		if err != nil {
			internalServerError(w, err)
			return
		}

		// save this to the datastore
		err = db.Update(func(tx *bolt.Tx) error {
			return rod.PutJson(tx, pasteBucketNameStr, paste.Id, paste)
		})
		if err != nil {
			internalServerError(w, err)
			return
		}

		// fmt.Printf("--> redirecting\n")
		http.Redirect(w, r, "/"+paste.Id, http.StatusFound)
	})

	m.Get("/:id", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vals(r)["id"]
		// fmt.Printf("id=%s\n", id)

		// See if this is for the paste page `/TtysPe` or the raw paste `/TtysPe.txt`.
		raw := strings.HasSuffix(id, ".txt")
		if raw {
			// remove the trailing ".txt" if this is a raw URL
			id = strings.TrimSuffix(id, ".txt")
		}

		// get the paste info from the datastore
		paste := Paste{}
		err := db.View(func(tx *bolt.Tx) error {
			return rod.GetJson(tx, pasteBucketNameStr, id, &paste)
		})
		if err != nil {
			internalServerError(w, err)
			return
		}

		// ToDo: check to see if this paste has expired
		if paste.Expire.IsZero() {
			// no expiry set
		} else {
			// check paste.Expire
		}

		// check if the file exists (even though it should)
		filename := dir + "/" + id
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			notFound(w, r)
			return
		}

		if raw {
			// open the file
			file, err := os.Open(filename)
			if err != nil {
				internalServerError(w, err)
				return
			}

			// write the plaintext header and stream the file
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, err = io.Copy(w, file)
			if err != nil {
				internalServerError(w, err)
				return
			}
			return
		}

		// rendering the paste page, so we're going to read in the file in it's entirety
		text, err := ioutil.ReadFile(filename)
		if err != nil {
			internalServerError(w, err)
			return
		}

		// render the Paste page
		data := struct {
			Apex    string
			BaseUrl string
			Paste   Paste
			Text    string
		}{
			apex,
			baseUrl,
			paste,
			string(text),
		}
		render(w, tmpl, "paste.html", data)
	})

	m.Get("/dl/:id", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vals(r)["id"]

		// check if the file exists (even though it should)
		filename := dir + "/" + id
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			notFound(w, r)
			return
		}

		// open the file
		file, err := os.Open(filename)
		if err != nil {
			internalServerError(w, err)
			return
		}

		// write the plaintext header and stream the file
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+id+".txt")
		_, err = io.Copy(w, file)
		if err != nil {
			internalServerError(w, err)
			return
		}
	})

	// finally, check all routing was added correctly
	check(m.Err)

	// server
	fmt.Printf("Starting server, listening on port %s\n", port)
	errServer := http.ListenAndServe(":"+port, m)
	check(errServer)
}
