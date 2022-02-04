package antireplay

import (
	"sync"
	"time"

	"github.com/seiflotfy/cuckoofilter"
)

const replayFilterCapacity = 100000

type ReplayFilter struct {
	lock     sync.Mutex
	poolA    *cuckoofilter.Filter
	poolB    *cuckoofilter.Filter
	poolSwap bool
	lastSwap int64
	interval int64
}

func NewReplayFilter(interval int64) *ReplayFilter {
	filter := &ReplayFilter{}
	filter.interval = interval
	return filter
}

func (filter *ReplayFilter) Interval() int64 {
	return filter.interval
}

func (filter *ReplayFilter) Check(sum []byte) bool {
	filter.lock.Lock()
	defer filter.lock.Unlock()

	now := time.Now().Unix()
	if filter.lastSwap == 0 {
		filter.lastSwap = now
		filter.poolA = cuckoofilter.NewFilter(replayFilterCapacity)
		filter.poolB = cuckoofilter.NewFilter(replayFilterCapacity)
	}

	elapsed := now - filter.lastSwap
	if elapsed >= filter.Interval() {
		if filter.poolSwap {
			filter.poolA.Reset()
		} else {
			filter.poolB.Reset()
		}
		filter.poolSwap = !filter.poolSwap
		filter.lastSwap = now
	}

	return filter.poolA.InsertUnique(sum) && filter.poolB.InsertUnique(sum)
}
