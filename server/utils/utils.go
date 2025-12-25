package utils

import (
	"math/rand"
	"time"
)


func GenerateRandomDuration(max ,min int) time.Duration{
	if min > max{
		panic("Invalid range of time")

	}
	rand.Seed(time.Now().UnixNano())
	radomSec :=rand.Intn(max-min+1) * int(time.Second)
	return time.Duration(radomSec) *time.Second
}