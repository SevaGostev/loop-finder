package analyze

type Align struct {
	Offset uint64
	Score  *CounterArray
}

type AlignQueue struct {
	Aligns []*Align
}

func (q *AlignQueue) Put(offset uint64, score *CounterArray) bool {

	for i := 0; i < len(q.Aligns); i++ {

		if score.Less(q.Aligns[i].Score) {
			
			var newAlign *Align

			if len(q.Aligns) == cap(q.Aligns) {
				newAlign = q.Aligns[len(q.Aligns)-1]
			} else {
				newAlign = &Align{Score: NewCounterArray(score.Length())}
				q.Aligns = append(q.Aligns, newAlign)
			}

			newAlign.Offset = offset
			newAlign.Score.CopyFrom(score)

			for j := len(q.Aligns) - 1; j > i; j-- {
				q.Aligns[j] = q.Aligns[j-1]
			}

			q.Aligns[i] = newAlign

			return true
		}
	}

	if len(q.Aligns) < cap(q.Aligns) {
		newScore := NewCounterArray(score.Length())
		newScore.CopyFrom(score)
		newAlign := &Align{Offset: offset, Score: newScore}
		q.Aligns = append(q.Aligns, newAlign)
		return true
	}

	return false
}

func (q *AlignQueue) WorstScore() *CounterArray {
	if len(q.Aligns) < cap(q.Aligns) {
		return nil
	} else {
		return q.Aligns[len(q.Aligns) - 1].Score
	}
}

func (q *AlignQueue) Length() int {
	return len(q.Aligns)
}

func (q *AlignQueue) Fill(offset uint64, score *CounterArray) {
	for _, a := range q.Aligns {
		a.Offset = offset
		a.Score.CopyFrom(score)
	}
}