package commands

// This file defines some variables to be available in all commands

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
