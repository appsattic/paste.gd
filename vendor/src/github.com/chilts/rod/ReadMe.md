# rod : helper functions inside BoltDB transactions

## Overview [![GoDoc](https://godoc.org/github.com/chilts/rod?status.svg)](https://godoc.org/github.com/chilts/rod) [![Build Status](https://travis-ci.org/chilts/rod.svg?branch=master)](https://travis-ci.org/chilts/rod) [![Code Climate](https://codeclimate.com/github/chilts/rod/badges/gpa.svg)](https://codeclimate.com/github/chilts/rod) [![Go Report Card](https://goreportcard.com/badge/github.com/chilts/rod)](https://goreportcard.com/report/github.com/chilts/rod)

Rod is a simple way to put and get values to/from a [BoltDB](https://github.com/boltdb/bolt) store. It can deal with
deep-hierarchies easily and is therefore a rod straight to the value you want.

## Installation

```sh
go get github.com/chilts/rod
```

Or (for `gb`):

```sh
gb vendor fetch github.com/chilts/rod
```

## Example ##

```go
user := User{
    Name: "chilts",
    Email: "andychilton@gmail.com",
    Logins: 1,
    Inserted: time.Now(),
}

db.Update(func(tx *bolt.TX) error {
    return rod.PutJson(tx, "users.chilts", "chilts", user)
})

```

## Author ##

By [Andrew Chilton](https://chilts.org/), [@twitter](https://twitter.com/andychilton).

For [AppsAttic](https://appsattic.com/), [@AppsAttic](https://twitter.com/AppsAttic).

## License ##

MIT.

(Ends)
