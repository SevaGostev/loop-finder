package analyze

import (
	"sync"

	"github.com/SevaGostev/loop-finder/data"
)

func FindBestAligns(arrA []data.SampleBuffer, arrB []data.SampleBuffer, sectionLength uint64, start uint64, end uint64, refStart uint64,
	numAligns uint, maxRoutines int, perRoutine uint64) *AlignQueue {

	counterCapacity := sectionLength / 2

	counterPool := sync.Pool {
		New: func() any {
			return NewCounterArray(counterCapacity)
		},
	}

	queuePool := sync.Pool {
		New: func() any {
			return &AlignQueue{Aligns: make([]*Align, 0, numAligns)}
		},
	}

	usedCounters := make([]*CounterArray, 0, maxRoutines)
	usedQueues := make([]*AlignQueue, 0, maxRoutines)

	out := AlignQueue{make([]*Align, 0, numAligns)}

	var wg sync.WaitGroup

	nextOffset := start

	for nextOffset < end {
		wg.Add(maxRoutines)

		routines := 0

		for i := 0; i < maxRoutines; i++ {

			if nextOffset >= end {
				break
			}

			endOffset := nextOffset + perRoutine
			if endOffset > end {
				endOffset = end
			}

			ctr := counterPool.Get().(*CounterArray)
			q := queuePool.Get().(*AlignQueue)

			go func(nxtO uint64, endO uint64, counter *CounterArray, best *AlignQueue) {
				compareRange(arrA, nxtO, endO, arrB, refStart, sectionLength, counter, best)
				wg.Done()
			}(nextOffset, endOffset, ctr, q)

			routines++
			usedCounters = append(usedCounters, ctr)
			usedQueues = append(usedQueues, q)

			nextOffset = endOffset
		}

		wg.Add(routines - maxRoutines)

		wg.Wait()
		
		for _, q := range usedQueues {

			for _, a := range q.Aligns {
				if !out.Put(a.Offset, a.Score) {
					break
				}
			}

			if out.WorstScore() != nil {
				q.Fill(0, out.WorstScore())
			}
			queuePool.Put(q)
		}
		
		usedQueues = usedQueues[0:0]

		for _, ctr := range usedCounters {
			ctr.Null()
			counterPool.Put(ctr)
		}

		usedCounters = usedCounters[0:0]
	}
	
	return &out
}