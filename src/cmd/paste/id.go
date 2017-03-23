package main

import (
	"math/rand"
	"time"
)

const idChars string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const idCharLen = len(idChars)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Id(len int) string {
	str := ""
	for i := 0; i < len; i++ {
		r := rand.Intn(idCharLen)
		str = str + string(idChars[r])
	}
	return str
}
