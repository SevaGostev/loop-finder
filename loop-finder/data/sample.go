package data

import (
	"math"
)

//Samples are in reverse order
type DecodedFile struct {
	
	Samples []SampleBuffer
	SampleRate int
}

type SampleBuffer interface {
	LessThan(uint64, SampleBuffer, uint64) bool
	GreaterThan(uint64, SampleBuffer, uint64) bool
	Diff(uint64, SampleBuffer, uint64) uint64
	SetMaxSample(uint64)
	SetMinSample(uint64)
	SetNeutralSample(uint64)
	Set(uint64, any)
	Get(uint64) interface{}
	Length() uint64
	Sub(uint64, uint64) SampleBuffer
}

type SampleBufferU16 struct {
	buffer []uint16
}

func (s SampleBufferU16) LessThan(offA uint64, bufferB SampleBuffer, offB uint64) bool {

	valB, ok := bufferB.Get(offB).(uint16)

	if !ok {
		return false
	}

	return s.buffer[offA] < valB
}

func (s SampleBufferU16) GreaterThan(offA uint64, bufferB SampleBuffer, offB uint64) bool {

	valB, ok := bufferB.Get(offB).(uint16)

	if !ok {
		return false
	}

	return s.buffer[offA] > valB
}

func (s SampleBufferU16) Diff(offA uint64, bufferB SampleBuffer, offB uint64) uint64 {

	valB, ok := bufferB.Get(offB).(uint16)

	if !ok {
		return uint64(0)
	}

	d := math.Pow( float64(s.buffer[offA] - valB), 2) * math.MaxFloat64

	return uint64(d)
}

func (s SampleBufferU16) SetMaxSample(offset uint64) {
	s.buffer[offset] = math.MaxUint16
}

func (s SampleBufferU16) SetMinSample(offset uint64) {
	s.buffer[offset] = 0
}

func (s SampleBufferU16) SetNeutralSample(offset uint64) {
	s.buffer[offset] = math.MaxUint16 / 2 + 1
}

func (s SampleBufferU16) Set(offset uint64, val interface{}) {
	v, ok := val.(uint16)

	if ok {
		s.buffer[offset] = v
	}
}

func (s SampleBufferU16) Get(offset uint64) interface{} {
	return s.buffer[offset]
}

func (s SampleBufferU16) Length() uint64 {
	return uint64(len(s.buffer))
}

func (s SampleBufferU16) Sub(from uint64, to uint64) SampleBuffer {
	return &SampleBufferU16{s.buffer[from:to]}
}



type SampleBufferF32 struct {
	buffer []float32
}

func (s SampleBufferF32) LessThan(offA uint64, bufferB SampleBuffer, offB uint64) bool {

	valB, ok := bufferB.Get(offB).(float32)

	if !ok {
		return false
	}

	return s.buffer[offA] < valB
}

func (s SampleBufferF32) GreaterThan(offA uint64, bufferB SampleBuffer, offB uint64) bool {

	valB, ok := bufferB.Get(offB).(float32)

	if !ok {
		return false
	}

	return s.buffer[offA] > valB
}

func (s SampleBufferF32) Diff(offA uint64, bufferB SampleBuffer, offB uint64) uint64 {

	valB, ok := bufferB.Get(offB).(float32)

	if !ok {
		return uint64(0)
	}

	d := math.Pow( float64(s.buffer[offA] - valB), 2) * math.MaxInt64
	return uint64(d)
}

func (s SampleBufferF32) SetMaxSample(offset uint64) {
	s.buffer[offset] = math.MaxFloat32
}

func (s SampleBufferF32) SetMinSample(offset uint64) {
	s.buffer[offset] = -math.MaxFloat32
}

func (s SampleBufferF32) SetNeutralSample(offset uint64) {
	s.buffer[offset] = float32(0)
}

func (s SampleBufferF32) Set(offset uint64, val interface{}) {
	v, ok := val.(float32)

	if ok {
		s.buffer[offset] = v
	}
}

func (s SampleBufferF32) Get(offset uint64) interface{} {
	return s.buffer[offset]
}

func (s SampleBufferF32) Length() uint64 {
	return uint64(len(s.buffer))
}

func (s SampleBufferF32) Sub(from uint64, to uint64) SampleBuffer {
	return &SampleBufferF32{s.buffer[from:to]}
}

func NewSampleBufferF32(length uint64) SampleBuffer {
	return &SampleBufferF32{make([]float32, length)}
}