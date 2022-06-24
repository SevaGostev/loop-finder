package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/SevaGostev/loop-finder/analyze"
	"github.com/SevaGostev/loop-finder/data"
	"github.com/SevaGostev/loop-finder/vorbis"
)

func handleFile(fn string, hint uint64) bool {

	name := filepath.Base(fn)

	var err error
	var f *os.File

	f, err = os.Open(fn)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Printf("File does not exist: %s\n", name)
		} else {
			log.Printf("Could not open file: %s\n", name)
		}

		return false
	}

	var nameNoExt, ext string
	var opts map[string]int
	nameNoExt, ext, opts, err = parseFilenameOptions(fn)

	if err != nil {
		log.Printf("Could not parse filename: %s\nError: %v\n", name, err)
		return false
	}

	if hint == 0 {
		ll, ok := opts["l"]
		
		if ok {
			if ll == 0 {
				log.Printf("Skipped %s (declared to loop end-to-end)\n", name)
				return false
			} else if ll > 0 {
				hint = uint64(ll)
			}
		}
	}

	switch ext {
		case ".ogg" :

			ls, ll, oldll, err := handleOgg(f, hint)

			if err != nil {
				log.Printf("Could not analyze file: %s\nError: %v\n", name, err)
				return false
			}

			if oldll == 0 {
				fmt.Printf("%s__s%dl%d\n", nameNoExt, ls, ll)
			} else {
				fmt.Printf("%s__s%dl%d  (%d)\n", nameNoExt, ls, ll, oldll)
			}
			return true

		default:
			log.Printf("Extension %s not supported: %s\n", ext, name)
			return false
		
	}
}

func parseFilenameOptions(fn string) (string, string, map[string]int, error) {

	pattern, err := regexp.Compile(`^([0-9A-Za-z\-]+(?:_[0-9A-Za-z]+)*)(?:__((?:[A-Za-z]\-?[0-9]*)+))?(\.[0-9A-Za-z]+)?$`)

	if err != nil {
		return "", "", nil, err
	}

	parts := pattern.FindStringSubmatch(fn)

	if parts == nil {
		return fn, "", make(map[string]int, 0), nil
	}

	name := parts[1]
	optstr := parts[2]
	ext := parts[3]

	if optstr == "" {
		return name, ext, make(map[string]int, 0), nil
	}

	optPattern, err := regexp.Compile(`(?:([A-Za-z])(\-?[0-9]*))`)

	if err != nil {
		return "", "", nil, err
	}

	opts := optPattern.FindAllStringSubmatch(optstr, -1)
	optMap := make(map[string]int, len(opts))

	for _, s := range opts {
		val, err := strconv.ParseInt(s[2], 10, 64)

		if err != nil {
			continue
		}

		optMap[s[1]] = int(val)
	}

	return name, ext, optMap, nil
}

func handleOgg(r *os.File, hint uint64) (uint64, uint64, uint64, error) {

	pcm, err := vorbis.Decode(r)

	if err != nil {
		return 0, 0, 0, err
	}

	oldll := uint64(0)
	oldllStr, ok := pcm.Comments["LOOPLENGTH"]

	if ok {
		var pErr error
		oldll, pErr = strconv.ParseUint(oldllStr, 10, 64)

		if pErr != nil {
			oldll = uint64(0)
		} else if oldll == 0 {
			return 0, 0, 0, errors.New("LOOPLENGTH is declared 0 (end-to-end loop)")
		}
	}

	if hint == 0 && oldll != 0 {
		hint = oldll
	}

	ls, ll, er := getLoopData(pcm, hint, 8)

	return ls, ll, oldll, er
}

func getLoopData(pcm *data.DecodedFile, hint uint64, maxRoutines int) (uint64, uint64, error) {

	stage0 := make([]data.SampleBuffer, len(pcm.Samples)*2)

	for i := range pcm.Samples {
		stage0[i * 2] = pcm.Samples[i]
		stage0[i * 2 + 1] = pcm.Samples[i]
	}
	
	stage1 := getQuantized(stage0, 16, maxRoutines)
	
	stage2 := getQuantized(stage1, 16, maxRoutines)

	sectionLength := uint64(pcm.SampleRate)

	var start, end uint64
	if hint != 0 {
		start = sectionLength
		if hint - sectionLength > start {
			start = hint - sectionLength
		}

		end = stage0[0].Length() - sectionLength
		if hint + sectionLength < end {
			end = hint + sectionLength
		}
	} else {
		start = stage0[0].Length() - (stage0[0].Length() / 4)
		end = stage0[0].Length() - sectionLength
	}

	if start >= end {
		return 0, 0, errors.New("file is too short")
	}

	best := analyze.AlignQueue{Aligns: make([]*analyze.Align, 0, 1)}
	aligns := analyze.FindBestAligns(stage2, stage2, sectionLength / 256, start / 256, end / 256, 0, 5, maxRoutines, 2048)
	
	for _, a := range aligns.Aligns {

		from := uint64(0)
		if a.Offset > 0 {
			from = (a.Offset - 1) * 16
		}

		to := (a.Offset + 1) * 16
		if to > stage1[0].Length() {
			to = stage1[0].Length()
		}
		
		align1 := analyze.FindBestAligns(stage1, stage1, sectionLength / 16, from, to, 0, 1, maxRoutines, 2048).Aligns[0]


		from = uint64(0)
		if align1.Offset > 0 {
			from = (align1.Offset - 1) * 16
		}

		to = (align1.Offset + 1) * 16
		if to > stage0[0].Length() {
			to = stage0[0].Length()
		}

		align0 := analyze.FindBestAligns(stage0, stage0, sectionLength, from, to, 0, 1, maxRoutines, 2048).Aligns[0]

		best.Put(align0.Offset, align0.Score)
	}

	loopLength := best.Aligns[0].Offset

	neutralSample := data.NewSampleBufferF32(1)
	neutralSample.SetNeutralSample(0)

	loopStart := uint64(0)
	loopStartScore := math.MaxFloat64

	for i := uint64(0); i < sectionLength && i + loopLength < stage0[0].Length(); i++ {

		score := float64(0)

		for _, c := range pcm.Samples {
			score += float64(c.Diff(i, c, i + loopLength) * 2)
			score += float64(c.Diff(i, neutralSample, 0))
		}

		if score < loopStartScore {
			loopStartScore = score
			loopStart = i + loopLength
		}
	}

	return uint64(stage0[0].Length() - loopStart), uint64(loopLength), nil
}


func getQuantized(in []data.SampleBuffer, blockSize int, maxRoutines int) []data.SampleBuffer {

	out := make([]data.SampleBuffer, 0, len(in))

	for i := 0; i < len(in); i += 2 {

		newLen := uint64( math.Ceil( float64(in[i].Length()) / float64(blockSize) ) )
		newMin := data.NewSampleBufferF32(newLen)
		newMax := data.NewSampleBufferF32(newLen)

		analyze.RequantizeParallel(in[i], in[i+1], newMin, newMax, blockSize, maxRoutines, 2048)

		out = append(out, newMin)
		out = append(out, newMax)
	}

	return out
}