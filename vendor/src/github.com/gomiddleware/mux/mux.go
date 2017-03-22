package mux

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"
)

type key int

const valsIdKey key = 999

// Errors that can be returned from this package.
var (
	// ErrPathMustStartWithSlash is returned when you create a route which doesn't start with a slash.
	ErrPathMustStartWithSlash = errors.New("mux: path must start with a slash")

	// ErrMultipleHandlers is returned when you create a route with multiple handlers.
	ErrMultipleHandlers = errors.New("mux: route has been given two handlers but only one can be provider")

	// ErrMiddlewareAfterHandler is returned when you create a route which has some middleware defined after the
	// handler.
	ErrMiddlewareAfterHandler = errors.New("mux: route can't have middleware defined after the handler")

	// ErrUnknownTypeInRoute is returned when something unexpected is passed to a route function.
	ErrUnknownTypeInRoute = errors.New("mux: unexpected type passed to route")
)

// Route is an internal "method + path + middlewares + handler" type created when each route is added. When adding a
// handler for Get(), Post(), Put(), Delete(), Options(), and Patch(), the middlewares prior to this route (and any on
// this route) are combined to create the final handler.
//
// These are not computed during routing but when added to the router, therefore they have negligible overhead.
type Route struct {
	Method      string
	Path        string
	Segments    []string
	Length      int
	Middlewares []func(http.Handler) http.Handler
	Handler     http.Handler
}

// Prefix is an internal "path + middlewares" type created when each middleware prefix is added. When adding we'll
// add the middlewares to the array of Middlewares.
type Prefix struct {
	Path        string
	Segments    []string
	Length      int
	Middlewares []func(http.Handler) http.Handler
	Handler     http.Handler
}

// Mux is an array of routes, prefixes, and an error (if one has happened).
type Mux struct {
	routes   []Route
	prefixes []Prefix
	Err      error
}

// Make sure the Mux conforms with the http.Handler interface.
var _ http.Handler = New()

// New returns a new initialized Mux.  Nothing is automatic. You must do slash/non-slash redirection yourself.
func New() *Mux {
	return &Mux{}
}

// Get is a shortcut for mux.add("GET", path, things...)
func (m *Mux) Get(path string, things ...interface{}) {
	m.add("GET", path, things...)
}

// Post is a shortcut for mux.add("POST", path, things...)
func (m *Mux) Post(path string, things ...interface{}) {
	m.add("POST", path, things...)
}

// Put is a shortcut for mux.add("PUT", path, things...)
func (m *Mux) Put(path string, things ...interface{}) {
	m.add("PUT", path, things...)
}

// Patch is a shortcut for mux.add("PATCH", path, things...)
func (m *Mux) Patch(path string, things ...interface{}) {
	m.add("PATCH", path, things...)
}

// Delete is a shortcut for mux.add("DELETE", path, things...)
func (m *Mux) Delete(path string, things ...interface{}) {
	m.add("DELETE", path, things...)
}

// Options is a shortcut for mux.add("OPTIONS", path, things...)
func (m *Mux) Options(path string, things ...interface{}) {
	m.add("OPTIONS", path, things...)
}

// Head is a shortcut for mux.add("HEAD", path, things...)
func (m *Mux) Head(path string, things ...interface{}) {
	m.add("HEAD", path, things...)
}

// Use adds some middleware to a path prefix. Unlike other methods such as Get, Post, Put, Patch, and Delete, Use
// matches for the prefix only and not the entire path. (Though of course, the entire exact path also matches.)
//
// e.g. m.Use("/profile/", ...) matches the requests "/profile/", "/profile/settings", and "/profile/a/path/to/".
//
// Note however, m.Use("/profile/", ...) doesn't match "/profile" since it contains too many slashes. But
// m.Use("/profile", ...) does match "/profile/" and "/profile/..." (but check that's actually what you want here).
func (m *Mux) Use(path string, things ...interface{}) {
	m.add("USE", path, things...)
}

// All adds a handler to a path prefix for all methods. Essentially a catch-all. Unlike other methods such as Get,
// Post, Put, Patch, and Delete, All matches for the prefix only and not the entire path.
//
// e.g. m.All("/s", ...) matches the requests "/s/img.png", "/s/css/styles.css", and "/s/js/app.js".
func (m *Mux) All(path string, things ...interface{}) {
	m.add("ALL", path, things...)
}

// add registers a new request handle with the given path and method.
//
// The respective shortcuts (for GET, POST, PUT, PATCH and DELETE) can also be used.
func (m *Mux) add(method, path string, things ...interface{}) {
	if m.Err != nil {
		return
	}

	if len(path) == 0 || path[0] != '/' {
		m.Err = ErrPathMustStartWithSlash
		return
	}

	if m.routes == nil {
		m.routes = make([]Route, 0)
	}

	// collect up some things like the middlewares and the handler
	var handler http.Handler
	var middlewares []func(http.Handler) http.Handler

	segments := strings.Split(path, "/")[1:]

	for _, thing := range things {
		switch val := thing.(type) {
		case func(http.Handler) http.Handler:
			// if we already have a handler, then we should bork
			if handler != nil {
				m.Err = ErrMiddlewareAfterHandler
				return
			}
			// all good, so add the middleware
			middlewares = append(middlewares, val)

		case func(http.HandlerFunc) http.HandlerFunc:
			// if we already have a handler, then we should bork
			if handler != nil {
				m.Err = ErrMiddlewareAfterHandler
				return
			}

			// This `setup` function is called when we are setting up our middleware stack. It is a `func(next
			// http.Handler) http.Handler` function as is the rest of the middleware.
			setup := func(next http.Handler) http.Handler {
				// We have been called here during middleware stack setup. So we need to return another `http.Handler`
				// which calls `next.ServeHTTP(w, r)`.
				//
				// To do this, call our original val() function with a `func(w http.ResponseWriter, r *http.Request)`)
				// so that we get a `func(w http.ResponseWriter, r *http.Request)` back. We convert that to a http.Handler()
				// so that it can be called from the previous middleware at the right time.
				//
				// Once `myNext` is called, that's actually the `val` middleware running and calling it's own next,
				// which is `myNext`, so we just pass that along to the `next` middleware we were given during setup
				// above.
				myNext := func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				}

				return http.Handler(val(myNext))
			}

			middlewares = append(middlewares, setup)

		case http.Handler:
			if handler != nil {
				m.Err = ErrMultipleHandlers
				return
			}
			// all good, so remember the handler
			handler = val

		case func(http.ResponseWriter, *http.Request):
			if handler != nil {
				m.Err = ErrMultipleHandlers
				return
			}
			// all good, so remember the handler
			handler = http.HandlerFunc(val)

		default:
			m.Err = ErrUnknownTypeInRoute
			return
		}
	}

	// If this is middleware, ie. USE, then there is nothing more to do, but if it is any other method, then we need to
	// create the final handler from any prefix middleware prior to this, and any middleware AND handler for this route.
	// If there is no handler for this route, then it is an error.
	if method == "USE" {
		if handler != nil {
			// this is not an error, since you might have a static server for a prefix, such as "/s"
		}
		prefix := Prefix{
			Path:        path,
			Segments:    segments,
			Length:      len(segments),
			Middlewares: middlewares,
			Handler:     handler,
		}

		// add  it to the middlewares
		m.prefixes = append(m.prefixes, prefix)
	} else {
		// GET, PUT, PATCH, POST, DELETE, OPTIONS, HEAD, and ALL!

		// generate our wrapped handler, wrapping each in reverse order from the current route, back down through each route
		wrappedHandler := handler
		for i := range middlewares {
			middleware := middlewares[len(middlewares)-1-i]
			wrappedHandler = middleware(wrappedHandler)
		}

		// now, go in reverse order through each added middleware and do the same thing
		for j := range m.prefixes {
			prefix := m.prefixes[len(m.prefixes)-1-j]

			if isPrefixMatch(segments, prefix.Segments) {
				// and again, get each middleware in reverse order
				for i := range prefix.Middlewares {
					middleware := prefix.Middlewares[len(prefix.Middlewares)-1-i]
					wrappedHandler = middleware(wrappedHandler)
				}
			}
		}

		// create our handler which contains everything we need
		route := Route{
			Method:      method,
			Path:        path,
			Segments:    segments,
			Length:      len(segments),
			Middlewares: nil, // we've already wrapped the handler
			Handler:     wrappedHandler,
		}

		// add it to the route handlers
		m.routes = append(m.routes, route)
	}
}

func isPrefixMatch(segments []string, prefixSegments []string) bool {
	prefixLength := len(prefixSegments)

	// if segments is just []string{''} (ie, from "/"), then this will match everything
	if prefixLength == 1 && prefixSegments[0] == "" {
		return true
	}

	// can't match if the prefix path length is longer than the URL
	if prefixLength > len(segments) {
		return false
	}

	// check each segment is the same (for the length of the prefix)
	for i, segment := range prefixSegments {
		// if both segments are empty, then this matches
		if segment == "" && segments[i] == "" {
			continue
		}

		// check if segment start with a ":"
		if segment[0:0] == ":" {
			continue
		}

		// check if this segment matches what is expected
		if segments[i] != segment {
			return false
		}
	}

	// nothing stopped us from matching, so it must be true
	return true
}

func isMatch(method string, segments []string, route *Route) (map[string]string, bool) {
	// can't match if the methods are different
	if route.Method != method {
		return nil, false
	}

	// can't match if the url length is different from the route length
	if route.Length != len(segments) {
		return nil, false
	}

	vals := make(map[string]string)

	// check each segment is the same (for the length of the prefix)
	for i, segment := range route.Segments {
		// if both segments are empty, then this matches
		if segment == "" && segments[i] == "" {
			continue
		}

		// check if segment start with a ":"
		if segment != "" && segment[0:1] == ":" {
			// store this segment into the vals[]
			vals[segment[1:]] = segments[i]
			continue
		}

		if segments[i] != segment {
			return nil, false
		}
	}

	// nothing stopped us from matching, so it must be true
	return vals, true
}

// ServeHTTP makes the router implement the http.Handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	normPath := path.Clean(r.URL.Path)

	// if the original path ends in a slash
	if normPath != "/" {
		if strings.HasSuffix(r.URL.Path, "/") {
			normPath = normPath + "/"
		}
	}

	// if these paths differ, then redirect to the real one
	if normPath != r.URL.Path {
		http.Redirect(w, r, normPath, http.StatusFound)
		return
	}

	// split on each segment (and discard the first blank one)
	segments := strings.Split(normPath, "/")[1:]

	for _, route := range m.routes {
		var vals map[string]string
		var matched bool

		// check if we need to check against a prefix or the entire path
		if route.Method == "ALL" {
			// check the prefix
			matched = isPrefixMatch(segments, route.Segments)
			if matched {
				vals = make(map[string]string)
				vals["path"] = strings.TrimPrefix(normPath, route.Path)
			}
		} else {
			// check the entire path
			vals, matched = isMatch(method, segments, &route)
		}
		if matched == false {
			continue
		}

		// save these placeholders into the context (even if empty)
		ctx := context.WithValue(r.Context(), valsIdKey, vals)
		r = r.WithContext(ctx)

		// and call the handler
		route.Handler.ServeHTTP(w, r)

		// nothing else to do, so stop multiple matches and multiple response.WriteHeader calls
		return
	}

	// If we got through to here, then not route matched, so just call NotFound.
	http.NotFound(w, r)
}

// Vals enables you to retrieve the placeholder matches of the current request.
func Vals(r *http.Request) map[string]string {
	return r.Context().Value(valsIdKey).(map[string]string)
}
