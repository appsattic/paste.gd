package main

import "time"

type Paste struct {
	Id         string
	Title      string
	Size       int
	Visibility string // public, unlisted, encrypted
	Expire     time.Time
	Created    time.Time
	Updated    time.Time
}
