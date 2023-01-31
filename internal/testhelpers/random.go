package testhelpers

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func RandomString() string {
	return strconv.Itoa(newRand().Intn(1000000))
}

func RandomError() error {
	return fmt.Errorf("random error! %s", RandomString())
}

func RandomInt() int {
	return newRand().Intn(1000000)
}

func RandomIntWithin(low, high int) int {
	return low + rand.Intn(high-low)
}

func RandomUInt() uint64 {
	return uint64(newRand().Int63n(1000000))
}

func RandomFloat() float32 {
	return newRand().Float32()
}

func RandomFloat64() float64 {
	return newRand().Float64()
}

func RandomTimeDuration() time.Duration {
	return time.Duration(newRand().Int())
}

func newRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
