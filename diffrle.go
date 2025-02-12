package diffrle

import (
	"fmt"
	"log"

	"github.com/tidwall/btree"
)

// for items of len 1 step doesn't really matter,
// but having some non-0 number helps to not worry about zeroes
const defaultStep = 1

// Seq is a struct to store range (FirstID is a key in btree)
type Seq struct {
	Step  int64
	Count int64
}

// Range of sequential values, evently spaced by step
type Range struct {
	FirstID int64
	Step    int64
	Count   int64
}

func (r Range) String() string {
	return fmt.Sprintf("<%d|%d:%d|%d>", r.FirstID, r.Step, r.Count, r.LastID())
}

// LastID returns last id of the range
func (r Range) LastID() int64 {
	if r.Step == 0 {
		return r.FirstID
	}
	return r.FirstID + (r.Count-1)*r.Step
}

// ContainsID return true if id is in this range
func (r Range) ContainsID(id int64) bool {
	if r.Step == 0 {
		return id == r.FirstID
	}
	return id >= r.FirstID && id <= r.LastID() && (id-r.FirstID)%r.Step == 0
}

// Set that is optimized for storing sequential IDs
type Set struct {
	m *btree.Map[int64, Seq]
}

// NewSet returns empty set
func NewSet(degree int) *Set {
	return &Set{
		m: btree.NewMap[int64, Seq](degree),
	}
}

// NewSetFromRanges returns set initialized with ranges
func NewSetFromRanges(degree int, rr []Range) *Set {
	m := btree.NewMap[int64, Seq](degree)
	for _, v := range rr {
		m.Set(v.FirstID, Seq{
			Count: v.Count,
			Step:  v.Step,
		})
	}
	return &Set{
		m: m,
	}
}

// Exists returns true if ID exists in set
func (s Set) Exists(id int64) bool {
	var prev *Range
	s.m.Descend(id, func(key int64, df Seq) bool {
		prev = &Range{
			FirstID: key,
			Step:    df.Step,
			Count:   df.Count,
		}
		return false
	})
	if prev == nil {
		return false
	}
	return prev.ContainsID(id)
}

func (s Set) setRange(r *Range) {
	if r.Count <= 0 {
		panic("wrong logic - count <= 0")
	}
	if r.Step <= 0 {
		panic("wrong logic - step < =0")
	}
	s.m.Set(r.FirstID, Seq{
		Step:  r.Step,
		Count: r.Count,
	})
}

// Compacts items around provided key
func (s Set) compactAdjacentRanges(r *Range) {
	if r == nil {
		return
	}

	// compact ranges to the left
	for {
		var prev *Range
		s.m.Descend(r.FirstID-1, func(key int64, df Seq) bool {
			prev = &Range{
				FirstID: key,
				Step:    df.Step,
				Count:   df.Count,
			}
			return false
		})
		merged := s.compactRanges(prev, r)
		if merged != nil {
			r = merged
		}
		if merged == nil {
			break
		}
	}

	// compact ranges to the right
	for {
		var next *Range
		s.m.Ascend(r.LastID()+1, func(key int64, df Seq) bool {
			next = &Range{
				FirstID: key,
				Step:    df.Step,
				Count:   df.Count,
			}
			return false
		})
		merged := s.compactRanges(r, next)
		if merged != nil {
			r = merged
		}
		if merged == nil {
			break
		}
	}
}

// compacts ranges so that they take less space
// either 2 ranges are compacted into a single one
// or IDs are moved from smaller ranges to bigger ranges, allowing for smaller
// ranges to be merged into other ranges in future
func (s Set) compactRanges(r1 *Range, r2 *Range) *Range {
	if r1 == nil || r2 == nil {
		return nil
	}
	if r1.FirstID > r2.FirstID {
		panic("expect ordered items")
	}

	// compact 2 single ids into range
	if r1.Count == 1 && r2.Count == 1 {
		merged := &Range{
			FirstID: r1.FirstID,
			Step:    r2.FirstID - r1.FirstID,
			Count:   2,
		}
		s.m.Delete(r2.FirstID)
		s.setRange(merged)
		return merged
	}

	// items without count have no step, so we assume its based on other item that has it
	if r2.Count == 1 {
		r2.Step = r1.Step
	}
	if r1.Count == 1 {
		r1.Step = r2.Step
	}

	// if steps are same - 2 ranges that sit next to each other merged into a bigger range
	if r1.Step == r2.Step {
		if r1.LastID()+r1.Step != r2.FirstID {
			return nil
		}
		s.m.Delete(r2.FirstID)
		merged := &Range{
			FirstID: r1.FirstID,
			Step:    r1.Step,
			Count:   r1.Count + r2.Count,
		}
		s.setRange(merged)
		return merged
	}

	// steal item from smaller range
	// first it helps to consolidate ranges
	// second it avoids same id moving back and forth between 2 ranges
	if r2.Count > r1.Count {
		// steal item from left range
		if r1.LastID() == r2.FirstID-r2.Step {
			s.m.Delete(r2.FirstID)
			newR2 := &Range{
				FirstID: r2.FirstID - r2.Step,
				Step:    r2.Step,
				Count:   r2.Count + 1,
			}
			s.setRange(newR2)
			merged := &Range{
				FirstID: r1.FirstID,
				Step:    r1.Step,
				Count:   r1.Count - 1,
			}
			s.setRange(merged)
			return merged
		}
	} else {
		// steal item from right range
		if r1.LastID()+r1.Step == r2.FirstID {
			s.m.Delete(r2.FirstID)
			newR2 := &Range{
				FirstID: r2.FirstID + r2.Step,
				Step:    r2.Step,
				Count:   r2.Count - 1,
			}
			s.setRange(newR2)
			merged := &Range{
				FirstID: r1.FirstID,
				Step:    r1.Step,
				Count:   r1.Count + 1,
			}
			s.setRange(merged)
			return merged
		}
	}
	return nil
}

// Set adds new ID to set.
// If ID already exists - nothing is done, ensuring only unique values
func (s Set) Set(id int64) {
	inserted := s.addID(id)
	s.compactAdjacentRanges(inserted)

}

func (s Set) addID(id int64) *Range {
	var prev *Range
	s.m.Descend(id, func(key int64, df Seq) bool {
		if prev == nil {
			prev = &Range{
				FirstID: key,
				Step:    df.Step,
				Count:   df.Count,
			}
			return true
		}
		return false
	})

	if prev != nil {
		if prev.ContainsID(id) { // already exists
			return nil
		}
		if id > prev.FirstID && id < prev.LastID() { // cut sequence in half
			cutAfterIndex := (id - prev.FirstID) / prev.Step
			newPrevItems := cutAfterIndex + 1
			newPrev := &Range{
				FirstID: prev.FirstID,
				Step:    prev.Step,
				Count:   newPrevItems,
			}
			s.setRange(newPrev)

			newNext := &Range{
				FirstID: newPrev.LastID() + prev.Step,
				Step:    prev.Step,
				Count:   prev.Count - newPrevItems,
			}

			s.setRange(newNext)
			r := &Range{
				FirstID: id,
				Step:    defaultStep,
				Count:   1,
			}
			s.setRange(r)
			return r
		}
	}
	r := &Range{
		FirstID: id,
		Step:    defaultStep,
		Count:   1,
	}
	// no overlap with existing ranges - create new range with single item
	s.m.Set(r.FirstID, Seq{
		Step:  r.Step,
		Count: r.Count,
	})
	return r
}

// Delete ID from the set
// returns true if item was found and deleted
func (s Set) Delete(id int64) bool {
	var prev *Range
	s.m.Descend(id, func(key int64, df Seq) bool {
		if prev == nil {
			prev = &Range{
				FirstID: key,
				Step:    df.Step,
				Count:   df.Count,
			}
			return true
		}
		return false
	})

	if prev == nil {
		return false
	}
	if !prev.ContainsID(id) {
		return false
	}

	if prev.Count == 1 {
		s.m.Delete(id)
		return true
	}

	if id == prev.FirstID {
		s.m.Delete(prev.FirstID)
		newPrev := &Range{
			FirstID: prev.FirstID + prev.Step,
			Step:    prev.Step,
			Count:   prev.Count - 1,
		}
		s.setRange(newPrev)
		return true
	}

	if id == prev.LastID() {
		newPrev := &Range{
			FirstID: prev.FirstID,
			Step:    prev.Step,
			Count:   prev.Count - 1,
		}
		s.setRange(newPrev)
		return true
	}

	if id > prev.FirstID && id < prev.LastID() { // cut sequence in half
		cutAfterIndex := (id - prev.FirstID) / prev.Step
		newPrev := &Range{
			FirstID: prev.FirstID,
			Step:    prev.Step,
			Count:   cutAfterIndex,
		}
		s.setRange(newPrev)

		newNext := &Range{
			FirstID: newPrev.LastID() + prev.Step + prev.Step,
			Step:    prev.Step,
			Count:   prev.Count - cutAfterIndex - 1,
		}
		s.setRange(newNext)
		return true
	}
	return false
}

// Ranges returns all underlying ranges in the set in ascending order
func (s Set) Ranges() []Range {
	rr := []Range{}
	s.m.Ascend(0, func(key int64, value Seq) bool {
		rr = append(rr, Range{
			FirstID: key,
			Step:    value.Step,
			Count:   value.Count,
		})
		return true
	})
	return rr
}

// IterAll iterates all IDs in the set in ascending order
func (s Set) IterAll(f func(id int64) bool) {
	s.m.Ascend(0, func(key int64, value Seq) bool {
		for i := int64(0); i < value.Count; i++ {
			cont := f(key + i*value.Step)
			if !cont {
				return cont
			}
		}
		return true
	})
}

// IterAll iterates IDs in ascending order in specified range [from,to)
func (s Set) IterFromTo(from, to int64, f func(id int64) bool) {
	s.m.Ascend(from, func(key int64, value Seq) bool {
		for i := int64(0); i < value.Count; i++ {
			v := key + i*value.Step
			if v >= to {
				return false
			}
			cont := f(v)
			if !cont {
				return cont
			}
		}
		return true
	})
}

// DeleteFromTo deletes IDs in specified range [from,to)
func (s Set) DeleteFromTo(from, to int64) {
	log.Print("BEFORE ", s.Ranges())
	defer func() {
		log.Print("AFTER ", s.Ranges())
	}()
	rr := []Range{}
	// get all affected ranges
	s.m.Descend(to, func(key int64, value Seq) bool {
		r := Range{
			FirstID: key,
			Step:    value.Step,
			Count:   value.Count,
		}
		if r.LastID() < from {
			return false
		}
		rr = append(rr, r)
		return true
	})
	if len(rr) == 0 { // no ranges before "to"
		return
	}

	for _, r := range rr {
		// delete whole range (applicable for all ranges in range [1:len(rr)-1]
		// -------<         range        >---------
		//   from-----------------------------to
		if from <= r.FirstID && to >= r.LastID() {
			s.m.Delete(r.FirstID)
			continue
		}
		// no intersection
		// -------------------<range>---------
		//   from--------to
		// OR
		// ----<range>---------
		//   -------------from--------to
		if from < r.FirstID && to < r.FirstID {
			continue
		}
		if from > r.LastID() && to > r.LastID() {
			continue
		}

		fromIndex := (from - r.FirstID) / r.Step
		fromRemainder := (from - r.FirstID) % r.Step
		toIndex := (to - r.FirstID) / r.Step

		keepLeftCount := fromIndex + 1
		if fromRemainder == 0 {
			keepLeftCount--
		}
		keepRightCount := r.Count - toIndex - 1

		if from > r.FirstID && to < r.LastID() {
			if keepLeftCount+keepRightCount == r.Count {
				// ---<  1               100              200     >---------
				//            from---to
				// keep left and right part - deleted range lies between steps
				continue
			}
		}

		// -------<         range        >---------
		//                 from-------------...
		//        <  kept >
		if keepLeftCount > 0 { // keep left
			newPrev := &Range{
				FirstID: r.FirstID,
				Step:    r.Step,
				Count:   keepLeftCount,
			}
			s.setRange(newPrev)
		} else {
			s.m.Delete(r.FirstID)
		}

		// -------<         range        >---------
		//   ...------------to
		//                    <   kept   >
		if keepRightCount > 0 { // keep right
			log.Print("RIGHT ", keepRightCount)
			newPrev := &Range{
				FirstID: r.LastID() - r.Step*(keepRightCount-1),
				Step:    r.Step,
				Count:   keepRightCount,
			}
			s.setRange(newPrev)
		}
	}
}
