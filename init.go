package main

import (
	"math/rand"
	"time"
)

func initRand() {
	rand.Seed(time.Now().UnixNano())
}
