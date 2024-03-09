package utils

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomInt(min, max int) int {
	return min + seededRand.Intn(max-min+1)
}

func RandomInt32(min, max int) int32 {
	return int32(RandomInt(min, max))
}

func RandomFloat(min, max int) float64 {
	return float64(RandomInt(min, max)) + seededRand.Float64()
}

func RandString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[seededRand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandOrderID() uuid.UUID {
	return uuid.New()
}
