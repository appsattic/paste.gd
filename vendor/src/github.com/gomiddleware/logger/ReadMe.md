# GoMiddleware : Logger #

Middleware that logs a request.

* [Project](https://github.com/gomiddleware/logger)
* [GoDoc](https://godoc.org/github.com/gomiddleware/logger)

## Synopsis ##

```go
package main

import (
	"net/http"

	"github.com/gomiddleware/logger"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.URL.Path))
}

func main() {
	// create the logger middleware
	log := logger.New()

	// make the http.Hander and wrap it with the log middleware
	handle := http.HandlerFunc(handler)
	http.Handle("/", log(handle))

	http.ListenAndServe(":8080", nil)
}
```

When curling this server, you'll see the following in the logs:

```
2016/10/01 10:05:43.625813 --- GET /about/
2016/10/01 10:05:43.625844 200 GET /about/ 7 34.568µs
2016/10/01 10:05:44.835607 --- GET /
2016/10/01 10:05:44.835636 200 GET / 1 26.633µs
2016/10/01 10:05:50.301461 --- POST /sdf?this=that
2016/10/01 10:05:50.301517 200 POST /sdf?this=that 4 50.09µs
```

Note the incoming requests (which have `---` in their lines) followed by the response (which, in this case all have
200). Then the method, URL, bytes written and time elapsed.

## Author ##

Written by [Andrew Chilton](https://chilts.org/) for [Apps Attic Ltd](https://appsattic.com/).

## License ##

ISC.

(Ends)

