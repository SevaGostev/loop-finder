package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SevaGostev/loop-finder/analyze"
	"github.com/SevaGostev/loop-finder/data"
	"github.com/SevaGostev/loop-finder/vorbis"
)


func main () {

	if len(os.Args) < 2 {
		log.Fatal("Specify input files")
	}

	if os.Args[1] == "-i" {

		if len(os.Args) == 2 {
			interactive("")
		} else {
			if _, err := os.Stat(os.Args[2]); err != nil {
				interactive("")
			} else {
				interactive(os.Args[2])
			}
		}
	} else {
		nonInteractive()
	}
}

func interactive(baseDir string) {

	fmt.Print("Interactive mode\n")

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		input, err := r.ReadString('\n')

		if err != nil {
			break
		}

		input = strings.TrimSuffix(input, "\n")

		if input == "" || input == "q" {
			break
		}

		tokens := strings.Split(input, ",")

		if len(tokens) < 2 {
			handleFile(filepath.Join(baseDir, input), 0)
		} else {
			hint, err := strconv.ParseUint(tokens[1], 10, 64)

			if err != nil {
				fmt.Print("Could not parse hint.\n")
				continue
			}

			handleFile(filepath.Join(baseDir, tokens[0]), hint)
		}
	}

	os.Exit(0)
}

func nonInteractive() {

	for i := 1; i < len(os.Args); i++ {

		handleFile(os.Args[i], 0)
	}

	os.Exit(0)
}

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

	ext := filepath.Ext(fn)
	nameNoExt := name[:strings.LastIndexByte(name, '.')]

	if ext == ".ogg" {

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

	} else {
		log.Printf("File not supported: %s\n", name)
		return false
	}
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
		start = 0
		if(hint >= sectionLength) {
			start = hint - sectionLength
		}

		end = stage0[0].Length()
		if hint + sectionLength < stage0[0].Length() {
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