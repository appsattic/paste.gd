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
var publicBucketNameStr = "public"
var publicBucketName = []byte(publicBucketNameStr)

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
	dumpDir := os.Getenv("PASTE_DUMP_DIR")
	if dumpDir == "" {
		log.Fatal("Specify a dir to write dumpfiles to 'PASTE_DUMP_DIR'")
	}
	googleAnalytics := os.Getenv("PASTE_GOOGLE_ANALYTICS")

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

	// dump the DB every 15 mins
	go dumpEvery(db, time.Duration(15)*time.Minute, dumpDir)

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

		// let's get all of the public paste keys only
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(publicBucketName)
			if b == nil {
				// weird ... !
				return nil
			}

			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				fmt.Fprintf(w, "%s/%s\n", baseUrl, k)
			}

			return nil
		})
		if err != nil {
			internalServerError(w, err)
			return
		}

	})

	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			PageName        string
			Apex            string
			BaseUrl         string
			GoogleAnalytics string
			Paste           Paste
			Form            map[string]string
			Errors          map[string]string
		}{
			"index",
			apex,
			baseUrl,
			googleAnalytics,
			Paste{},
			nil,
			nil,
		}
		render(w, tmpl, "index.html", data)
	})

	m.Get("/about", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			PageName        string
			Apex            string
			BaseUrl         string
			GoogleAnalytics string
			Paste           Paste
		}{
			"about",
			apex,
			baseUrl,
			googleAnalytics,
			Paste{},
		}
		render(w, tmpl, "about.html", data)
	})

	m.Get("/paste", redirect("/"))
	m.Post("/paste", func(w http.ResponseWriter, r *http.Request) {
		// get the values of certain fields
		title := r.FormValue("Title")
		text := r.FormValue("Text")
		visibility := r.FormValue("Visibility")
		if visibility != "unlisted" && visibility != "public" && visibility != "encrypted" {
			// since this comes from a form, then someone is messing with it - don't give them any leeway
			internalServerError(w, errors.New("visibility was not an allowed option"))
			return
		}

		// check that the paste is not empty
		if text == "" {
			form := make(map[string]string)
			form["Title"] = title
			form["Text"] = text
			form["Visibility"] = visibility
			errors := make(map[string]string)
			errors["Text"] = "Provide some text"
			data := struct {
				PageName        string
				Apex            string
				BaseUrl         string
				GoogleAnalytics string
				Paste           Paste
				Form            map[string]string
				Errors          map[string]string
			}{
				"paste",
				apex,
				baseUrl,
				googleAnalytics,
				Paste{},
				form,
				errors,
			}
			render(w, tmpl, "index.html", data)
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
			// check if this is a public paste and add the name to the public bucket
			if paste.Visibility == "public" {
				rod.PutString(tx, publicBucketNameStr, paste.Id, now.Format("20060201-150405.000000000"))
			}

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
			PageName        string
			Apex            string
			BaseUrl         string
			GoogleAnalytics string
			Paste           Paste
			Text            string
		}{
			"paste",
			apex,
			baseUrl,
			googleAnalytics,
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

	m.Get("/iframe/:id", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vals(r)["id"]

		// check if the file exists (even though it should)
		filename := dir + "/" + id
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			notFound(w, r)
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
			Apex            string
			BaseUrl         string
			GoogleAnalytics string
			Id              string
			Text            string
		}{
			apex,
			baseUrl,
			googleAnalytics,
			id,
			string(text),
		}
		render(w, tmpl, "iframe.html", data)
	})

	// finally, check all routing was added correctly
	check(m.Err)

	// server
	fmt.Printf("Starting server, listening on port %s\n", port)
	errServer := http.ListenAndServe(":"+port, m)
	check(errServer)
}
