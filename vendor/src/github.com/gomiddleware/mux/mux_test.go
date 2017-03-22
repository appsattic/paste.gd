package mux

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// GetTestHandler returns a http.HandlerFunc for testing http middleware
func GetTestHandler() http.HandlerFunc {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		panic("test entered test handler, this should not happen")
	}
	return http.HandlerFunc(fn)
}

func TestSimple(t *testing.T) {
	tests := []struct {
		Desc       string
		Url        string
		StatusCode int
		Body       string
	}{
		{"home", "/", 200, "Home\n"},
		{"about", "/about", 200, "About\n"},
		{"user", "/u/chilts/", 200, "Hello, chilts!\n"},
	}

	m := New()

	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Home\n"))
	})

	m.Get("/about", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("About\n"))
	})

	m.Get("/u/:username/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vals := Vals(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, " + vals["username"] + "!\n"))
	}))

	ts := httptest.NewServer(m)
	defer ts.Close()

	fmt.Printf("url=%s\n", ts.URL)

	for _, tc := range tests {
		res, err := http.Get(ts.URL + tc.Url)
		check(err)

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		check(err)

		fmt.Printf("body=%s", body)

		if tc.StatusCode != res.StatusCode {
			log.Fatal("Incorrect status code : " + tc.Url)
		}
		if tc.Body != string(body) {
			log.Fatal("Incorrect body : " + tc.Url)
		}
	}
}
