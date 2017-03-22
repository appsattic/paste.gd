# mux : powerful router, with first-class middleware chains.

## Overview [![GoDoc](https://godoc.org/github.com/gomiddleware/mux?status.svg)](https://godoc.org/github.com/gomiddleware/mux) [![Build Status](https://travis-ci.org/gomiddleware/mux.svg)](https://travis-ci.org/gomiddleware/mux)

gomiddlware/mux aims to provide a neat interface for a mux with middleware included as first-class functions as well as
endpoint handlers. Middleware and handlers are the two primary ways to compose powerful response mechanisms to any web
request using your mux. Instead of using a separate library for chaining your middleware or having many intermediate
variables and chaining them, just use the functions you want wherever you want.

The features that mux boasts are all idomatic Go, such as:

* using the standard context package
* middleware defined as `func(http.Handler) http.Handler`
* handlers defined as `http.Handler` or `http.HandlerFunc`
* no external dependencies, just plain net/http
* everything is explicit - and is very much considered a feature (see below for the things left out)

Instead of focussing on pure-speed using a trie based router implementations, gomiddleware/mux instead focuses on
being both small yet powerful. Some of the main features of mux are features that have been left out, such as:

* no sub-routers or mouting other routers
* no automatic case-folding on paths
* no automatic slash/non-slash redirection
* no adding router values into things like r.URL (uses context instead)

The combination of just middleware and handlers these two things give you a very powerful composition system where you
compose middleware on prefixes and middleware chains on endpoints.

## Installation

```sh
go get github.com/gomiddleware/mux
```

## Usage / Example

```go
// new Mux
m := mux.New()

// log every request
m.Use("/", logger.New())

// serve a static set of files under "/s/"
m.All("/s", http.FileServer(http.Dir("./static")))

// every (non-static) request gets a 'X-Request-ID' request header
m.Use("/", reqid.RandomId)

// serve the /about page
m.Get("/about", aboutHandler)

// serve home, with one middleware specific to it's route
m.Get("/", incHomeHits, homeHandler)

// now check if adding any of the routes failed
if m.Err != nil {
    log.Fatal(m.Err)
}

// start the server
http.ListenAndServe(":8080", m)
```

## Author ##

By [Andrew Chilton](https://chilts.org/), [@twitter](https://twitter.com/andychilton).

For [AppsAttic](https://appsattic.com/), [@AppsAttic](https://twitter.com/AppsAttic).

## LICENSE

MIT.
