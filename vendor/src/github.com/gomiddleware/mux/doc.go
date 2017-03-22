// Package mux is a simple(ish) mux, however, it provides a few nice features such as easy use of middleware chains and
// a router which doesn't automatically look at only prefixes (like the Go built-in mux). Everything is explicit and
// nothing is done for you. ie. we don't automatically redirect to a slash or non-slash version of each path, you need
// to do that yourself (or use https://github.com/gomiddleware/slash to help).
//
// There are two fundamentals at work here, middleware and handlers. Middleware is expected to either: (1) do something
// with the request/response/context on the way in, call the `next` middleware and optionally do something on the way
// out, or (2) deal with the request completely and not call the `next` middleware at all. It is up to the middleware to decide this.
//
// An example of middleware calling next might be adding a requestId in to each request's Context.
//
// An example of middleware dealing with the request might be to check that a user is logged in, and if not, sending a
// redirect to "/", or send back a 405.
//
// For the purposes of this library (and in general), a middleware is defined as one of these:
//
// • func(http.Handler) http.Handler
//
// • func(http.HandlerFunc) http.HandlerFunc
//
// For the purposes of this library (and in general), a handler is defined as:
//
// • http.Handler
//
// • func(http.ResponseWriter, *http.Request)
//
// To add middleware to any prefix, use:
//
//     m.Use("/prefix", middlewares...)
//
// To add a handler to any path, use one of Get/Post/Put/Patch/Delete/Options/Head such as:
//
//     m.Get("/path", [middlewares...,] handler)
//
// Of course, you can only add one handler to each route, and it must be the last thing you pass in.
//
// Don't worry that you are passing one of 4 types in to each method since they are organised at the time they are
// added, and not during the request cycle. Only type assertions are used and not reflection.
//
// Example
//
//     m := mux.New()
//
//     m.Use("/", logger.New())
//
//     m.Get("/", func(w http.ResponseWriter, r *http.Request) {
//         w.WriteHeader(http.StatusOK)
//         w.Write([]byte("Home\n"))
//     })
//
//     m.Get("/my/", checkUserIsSignedIn, userHomeHandler)
//
//     if m.Err != nil {
//         // ... something went wrong with a route
//     }
//
//     log.Fatal(http.ListenAndServe(":8080", m))
//
// More information and examples can be found at http://chilts.org/2017/01/27/gomiddleware-mux
//
package mux
