package analyze

import (
	"github.com/SevaGostev/loop-finder/data"
)

func compare(arrA data.SampleBuffer, arrB data.SampleBuffer, bound *CounterArray, outArr *CounterArray) bool {

	off := uint64(0)
	end := arrA.Length()

	for off < end {

		outArr.Add(arrA.Diff(off, arrB, off))

		if bound != nil {
			if !outArr.Less(bound) {
				return true
			}
		}

		off++
	}

	return false
}

func compareChannels(arrA []data.SampleBuffer, arrB []data.SampleBuffer, bound *CounterArray, outArr *CounterArray) bool {

	for i := range arrA {

		bounded := compare(arrA[i], arrB[i], bound, outArr)

		if bounded {
			return true
		}
	}

	return false
}

func compareRange(arrA []data.SampleBuffer, startA uint64, endA uint64, arrB []data.SampleBuffer, offB uint64,
	length uint64, counter *CounterArray, q *AlignQueue) {

	a := startA

	slicedA := make([]data.SampleBuffer, len(arrA))
	slicedB := make([]data.SampleBuffer, len(arrB))

	for i := range arrB {
		slicedB[i] = arrB[i].Sub(offB, offB + length)
	}

	for a < endA {

		counter.Null()

		for i := range arrA {
			slicedA[i] = arrA[i].Sub(a, a + length)
		}

		bounded := compareChannels(slicedA, slicedB, q.WorstScore(), counter)

		if !bounded {
			q.Put(a, counter)
		}
		
		a++
	}
}