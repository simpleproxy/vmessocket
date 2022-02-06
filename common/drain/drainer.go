package drain

import (
	"io"

	"github.com/vmessocket/vmessocket/common/dice"
)

type BehaviorSeedLimitedDrainer struct {
	DrainSize int
}

type NopDrainer struct{}

func drainReadN(reader io.Reader, n int) error {
	_, err := io.CopyN(io.Discard, reader, int64(n))
	return err
}

func NewBehaviorSeedLimitedDrainer(behaviorSeed int64, drainFoundation, maxBaseDrainSize, maxRandDrain int) (Drainer, error) {
	behaviorRand := dice.NewDeterministicDice(behaviorSeed)
	BaseDrainSize := behaviorRand.Roll(maxBaseDrainSize)
	RandDrainMax := behaviorRand.Roll(maxRandDrain) + 1
	RandDrainRolled := dice.Roll(RandDrainMax)
	DrainSize := drainFoundation + BaseDrainSize + RandDrainRolled
	return &BehaviorSeedLimitedDrainer{DrainSize: DrainSize}, nil
}

func NewNopDrainer() Drainer {
	return &NopDrainer{}
}

func WithError(drainer Drainer, reader io.Reader, err error) error {
	drainErr := drainer.Drain(reader)
	if drainErr == nil {
		return err
	}
	return newError(drainErr).Base(err)
}

func (d *BehaviorSeedLimitedDrainer) AcknowledgeReceive(size int) {
	d.DrainSize -= size
}

func (n NopDrainer) AcknowledgeReceive(size int) {}

func (d *BehaviorSeedLimitedDrainer) Drain(reader io.Reader) error {
	if d.DrainSize > 0 {
		err := drainReadN(reader, d.DrainSize)
		if err == nil {
			return newError("drained connection")
		}
		return newError("unable to drain connection").Base(err)
	}
	return nil
}

func (n NopDrainer) Drain(reader io.Reader) error {
	return nil
}
