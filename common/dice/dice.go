package dice

import (
	"math/rand"
	"time"
)

func Roll(n int) int {
	if n == 1 {
		return 0
	}
	return rand.Intn(n)
}

func RollDeterministic(n int, seed int64) int {
	if n == 1 {
		return 0
	}
	return rand.New(rand.NewSource(seed)).Intn(n)
}

func RollUint16() uint16 {
	return uint16(rand.Int63() >> 47)
}

func RollUint64() uint64 {
	return rand.Uint64()
}

func NewDeterministicDice(seed int64) *DeterministicDice {
	return &DeterministicDice{rand.New(rand.NewSource(seed))}
}

type DeterministicDice struct {
	*rand.Rand
}

func (dd *DeterministicDice) Roll(n int) int {
	if n == 1 {
		return 0
	}
	return dd.Intn(n)
}

func init() {
	rand.Seed(time.Now().Unix())
}
