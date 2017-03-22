# logit : structured logger

## Overview [![GoDoc](https://godoc.org/github.com/gomiddleware/logit?status.svg)](https://godoc.org/github.com/gomiddleware/logit) [![Build Status](https://travis-ci.org/gomiddleware/logit.svg?branch=master)](https://travis-ci.org/gomiddleware/logit) [![Code Climate](https://codeclimate.com/github/gomiddleware/logit/badges/gpa.svg)](https://codeclimate.com/github/gomiddleware/logit) [![Go Report Card](https://goreportcard.com/badge/github.com/gomiddleware/logit)](https://goreportcard.com/report/github.com/gomiddleware/logit) [![Sourcegraph](https://sourcegraph.com/github.com/gomiddleware/logit/-/badge.svg)](https://sourcegraph.com/github.com/gomiddleware/logit?badge)

A structured logger with a simple interface. Why overcomplicate things? Use this instead.

## Install

```
go get github.com/gomiddleware/logit
```

## Example

```
func main() {
	fmt.Println("Started")
	defer fmt.Println("Ended")

	log := logit.New(os.Stdout, "main")
	log.Output("Started")
	defer log.Output("Ended")

	log2 := log.Clone("datastore")
	log2.WithField("id", "deadbeef")
	log2.Output("started transaction")
	defer log2.Output("Ended transaction")

	log3 := log2.Clone("user")
	log3.WithField("user", "chilts")
	log3.Output("Found User")

	// time=20170313-040404.727880393 sys=main msg=Started
	// time=20170313-040404.727897420 sys=main.datastore id=deadbeef msg=started transaction
	// time=20170313-040404.727906222 sys=main.datastore.user id=deadbeef user=chilts msg=Found User
	// time=20170313-040404.727915401 sys=main.datastore id=deadbeef msg=Ended transaction
	// time=20170313-040404.727921160 sys=main msg=Ended
}
```

## Author ##

By [Andrew Chilton](https://chilts.org/), [@twitter](https://twitter.com/andychilton).

For [AppsAttic](https://appsattic.com/), [@AppsAttic](https://twitter.com/AppsAttic).

## LICENSE

MIT.
