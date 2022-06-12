package vorbis

import (
	"errors"
	"io"
	"strings"

	"github.com/SevaGostev/loop-finder/data"
	"github.com/jfreymuth/oggvorbis"
)


func Decode(r io.ReadSeeker) (*data.DecodedFile, error) {

	dec, err := oggvorbis.NewReader(r)

	if err != nil {
		return nil, err
	}
	
	

	numChannels := dec.Channels()
	length := dec.Length()
	coms := dec.CommentHeader()

	if length == 0 {
		return nil, errors.New("could not determine length")
	}

	out := data.DecodedFile{
		Samples: make([]data.SampleBuffer, numChannels),
		SampleRate: dec.SampleRate(),
		Comments: make(map[string]string, len(coms.Comments)),
	}

	for _, com := range coms.Comments {
		tokenized := strings.Split(com, "=")

		if len(tokenized) < 2 {
			continue
		}

		out.Comments[tokenized[0]] = tokenized[1]
	}

	for i := range out.Samples {
		out.Samples[i] = data.NewSampleBufferF32(uint64(length))
	}

	read := int64(0)
	end := int64(length) * int64(numChannels)
	buf := make([]float32, 256)
	off := uint64(length - 1)

	for {
		trbuf := buf[:len(buf) - (len(buf) % numChannels)]
		n, err := dec.Read(trbuf)

		read += int64(n)

		for i := 0; i < n; i += numChannels {

			for j := 0; j < numChannels; j++ {

				out.Samples[j].Set(off, trbuf[i + j])
			}

			off--
		}

		if err != nil {

			if !errors.Is(err, io.EOF) {
				return nil, err
			}

			break
		}

		if read >= end {
			break
		}
		if n == 0 {
			return nil, io.ErrNoProgress
		}
	}

	return &out, nil
}



