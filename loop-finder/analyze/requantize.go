package analyze

import (
	"math"
	"sync"

	"github.com/SevaGostev/loop-finder/data"
)

func requantize(minSamples data.SampleBuffer, maxSamples data.SampleBuffer, targetMinSamples data.SampleBuffer, targetMaxSamples data.SampleBuffer, blockLength int) {

	i := uint64(0)
	b := uint64(0)

	end := minSamples.Length()
	for i < end {
		targetMinSamples.SetMaxSample(b)
		targetMaxSamples.SetMinSample(b)

		blockEnd := blockLength
		if blockEnd > int(end - i) {
			blockEnd = int(end - i)
		}

		for j := 0; j < blockEnd; j++ {

			if minSamples.LessThan(i, targetMinSamples, b) {
				targetMinSamples.Set(b, minSamples.Get(i))
			}

			if maxSamples.GreaterThan(i, targetMaxSamples, b) {
				targetMaxSamples.Set(b, maxSamples.Get(i))
			}

			i++
		}

		b++
	}
}

func RequantizeParallel(minSamples data.SampleBuffer, maxSamples data.SampleBuffer, targetMinSamples data.SampleBuffer, targetMaxSamples data.SampleBuffer,
	blockLength int, maxRoutines int, perRoutine int) {

	var wg sync.WaitGroup

	nextOffset := uint64(0)
	nextTargetOffset := uint64(0)

	targetPerRoutine := uint64(math.Ceil(float64(perRoutine) / float64(blockLength)))
	end := minSamples.Length()
	targetEnd := targetMinSamples.Length()

	for nextOffset < end {
		wg.Add(maxRoutines)

		routines := 0

		for i := 0; i < maxRoutines; i++ {

			if nextOffset >= end {
				break
			}

			endOffset := nextOffset + uint64(perRoutine)
			if endOffset > end {
				endOffset = end
			}

			endTargetOffset := nextTargetOffset + targetPerRoutine
			if endTargetOffset > targetEnd {
				endTargetOffset = targetEnd
			}

			go func(minIn data.SampleBuffer, maxIn data.SampleBuffer, minOut data.SampleBuffer, maxOut data.SampleBuffer) {
				requantize(minIn, maxIn, minOut, maxOut, blockLength)
				wg.Done()
			}(minSamples.Sub(nextOffset, endOffset), maxSamples.Sub(nextOffset, endOffset),
				targetMinSamples.Sub(nextTargetOffset, endTargetOffset), targetMaxSamples.Sub(nextTargetOffset, endTargetOffset))

			routines++

			nextOffset = endOffset
			nextTargetOffset = endTargetOffset
		}

		wg.Add(routines - maxRoutines)

		wg.Wait()
	}
}