package protocol

import (
	"time"

	"github.com/vmessocket/vmessocket/common/dice"
)
type (
	Timestamp int64
	TimestampGenerator func() Timestamp
)

func NewTimestampGenerator(base Timestamp, delta int) TimestampGenerator {
	return func() Timestamp {
		rangeInDelta := dice.Roll(delta*2) - delta
		return base + Timestamp(rangeInDelta)
	}
}

func NowTime() Timestamp {
	return Timestamp(time.Now().Unix())
}
