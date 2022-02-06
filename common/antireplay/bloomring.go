package antireplay

import (
	"sync"

	"github.com/v2fly/ss-bloomring"
)

type BloomRing struct {
	*ss_bloomring.BloomRing
	lock *sync.Mutex
}

func NewBloomRing() BloomRing {
	const (
		DefaultSFCapacity = 1e6
		DefaultSFFPR  = 1e-6
		DefaultSFSlot = 10
	)
	return BloomRing{ss_bloomring.NewBloomRing(DefaultSFSlot, DefaultSFCapacity, DefaultSFFPR), &sync.Mutex{}}
}

func (b BloomRing) Check(sum []byte) bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.Test(sum) {
		return false
	}
	b.Add(sum)
	return true
}

func (b BloomRing) Interval() int64 {
	return 9999999
}
