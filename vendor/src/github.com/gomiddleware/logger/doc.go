// Package logger provides a logger which can be used as Go Middleware. It logs a line on an incoming request and
// another line when that request has finished, along with the final status code and the time elapsed.
//
// This package is inspured heavily by https://godoc.org/github.com/gohttp/logger but removes the use of
// github.com/dustin/go-humanize and github.com/segmentio/go-log, and instead uses a non-humanize response size
// (ie. plain byte-size). It uses gomiddleware/logit (a very small pkg) for structured logging.
//
// Currently there is only one type of logger, but more are likely to be added in the future.
//
package logger
