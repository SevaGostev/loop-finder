package analyze

import (
	"fmt"
	"math"
)

type CounterArray struct {
	arr []uint64
	pos int
}


func (a *CounterArray) Add(x uint64) {
	p := len(a.arr) - 1
	
	for p >= 0 {
		inv := math.MaxUint64 - a.arr[p]

		if inv >= x {
			a.arr[p] = a.arr[p] + x
			break
		}
		
		a.arr[p] = x - inv - 1
		x = 1
		p --
	}

	if a.pos > p {
		a.pos = p
	}

	if p < 0 {
		newArr := make([]uint64, len(a.arr), len(a.arr) * 2)
		newArr[len(a.arr) - 1] = x
		a.pos = len(a.arr) - 1
		newArr = append(newArr, a.arr...)
		a.arr = newArr
		
	}
}

func (a *CounterArray) Less(other *CounterArray) bool {
	left := a.arr
	right := other.arr

	if len(left) - a.pos > len(right) - other.pos {
		return false
	}

	if len(left) - a.pos < len(right) - other.pos {
		return true
	}

	lpos := a.pos
	rpos := other.pos

	for lpos < len(left) {

		if left[lpos] < right[rpos] {
			return true
		} else if left[lpos] > right[rpos] {
			return false
		}
		lpos++
		rpos++
	}

	return false
}

func (a *CounterArray) Length() uint64 {
	return uint64(len(a.arr))
}

func (a *CounterArray) CopyFrom(src *CounterArray) {
	if len(a.arr) < len(src.arr) {
		a.arr = make([]uint64, len(src.arr))
	}

	diff := len(a.arr) - len(src.arr)

	for i := 0; i < diff; i++ {
		a.arr[i] = 0
	}

	a.arr = append(a.arr[:diff], src.arr...)
	a.pos = src.pos + diff
}

func (a *CounterArray) Null() {
	for i := range a.arr {
		a.arr[i] = 0
	}
	a.pos = len(a.arr)
}

func (a *CounterArray) Print() {
	i := 0

	for i < len(a.arr) {
		if a.arr[i] != 0 {
			break
		}
		i ++
	}
	zeros := 0
	for i < len(a.arr) {
		if a.arr[i] != 0 {
			if zeros > 0 {
				fmt.Printf("%d-%d:0\n", i - zeros, i - 1)
				zeros = 0
			}
			fmt.Printf("%d:%d\n", i, a.arr[i])
		} else {
			zeros ++
		}
		i ++
	}
}

func NewCounterArray(length uint64) *CounterArray {
	return &CounterArray{arr: make([]uint64, length), pos: int(length)}
}