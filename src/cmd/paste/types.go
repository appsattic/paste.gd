package main

import "time"

type Paste struct {
	Id         string
	Title      string
	Text       string
	Size       int
	Visibility string // public, unlisted (private)
	Expire     time.Time
	Created    time.Time
	Updated    time.Time
}
