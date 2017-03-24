package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
		visibility := "public" // ToDo: r.FormValue("visibility")

		// create the paste
		now := time.Now().UTC()
		paste := Paste{
			Id:         Id(6),
			Title:      title,
			Text:       text,
			Size:       len(text),
			Visibility: visibility,
			Created:    now,
			Updated:    now,
		}
		fmt.Printf("--> paste=%#v\n", paste)

		// check visibility
		if visibility != "public" && visibility != "unlisted" {
			fmt.Printf("--> visibility=%#v\n", visibility)
			paste.Visibility = "public"

			data := struct {
				Apex    string
				BaseUrl string
				Paste   Paste
			}{
				apex,
				baseUrl,
				paste,
			}
			render(w, tmpl, "index.html", data)
			return
		}

		// save the text to a file
		filename := dir + "/" + paste.Id
		fmt.Printf("--> filename=%#v\n", filename)
		err = ioutil.WriteFile(filename, []byte(paste.Text), 0755)
		if err != nil {
			internalServerError(w, err)
			return
		}
		paste.Text = ""

		// save this to the datastore
		err = db.Update(func(tx *bolt.Tx) error {
			return rod.PutJson(tx, pasteBucketNameStr, paste.Id, paste)
		})
		if err != nil {
			internalServerError(w, err)
			return
		}

		fmt.Printf("--> redirecting\n")
		http.Redirect(w, r, "/"+paste.Id, http.StatusFound)
	})

	m.Get("/:id", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vals(r)["id"]
		fmt.Printf("id=%s\n", id)

		// No need to check the datastore, since if the person has the correct Id, then
		// they are allowed to see the paste.

		// Check the datastore first for existance first, probably quicker.

		// check if the file exists
		filename := dir + "/" + id
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			notFound(w, r)
			return
		}

		// open the file and stream it to the response
		file, err := os.Open(filename)
		if err != nil {
			internalServerError(w, err)
			return
		}

		// now write the plaintext header and stream the file
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err = io.Copy(w, file)
		if err != nil {
			internalServerError(w, err)
			return
		}
	})

	m.Get("/raw/:id", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vals(r)["id"]
		fmt.Printf("id=%s\n", id)

		// No need to check the datastore, since if the person has the correct Id, then
		// they are allowed to see the paste.

		// check if the file exists
		filename := dir + "/" + id
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			notFound(w, r)
			return
		}

		// open the file and stream it to the response
		file, err := os.Open(filename)
		if err != nil {
			internalServerError(w, err)
			return
		}

		// now write the plaintext header and stream the file
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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
