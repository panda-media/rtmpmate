package TimeRanges

import (
	"container/list"
	"rtmpmate.com/muxer/TimeRanges/TimeRange"
)

type TimeRanges struct {
	Ranges list.List
}

func New() (*TimeRanges, error) {
	tr, _ := TimeRange.New()

	var trs TimeRanges
	trs.Ranges.PushBack(tr)
	return &trs, nil
}

func (this *TimeRanges) Add(start float64, end float64) error {
	done := false

	for e := this.Ranges.Back(); e != nil; e = e.Prev() {
		v := e.Value.(*TimeRange.TimeRange)
		if start < v.Start && v.Start <= end {
			v.Start = start
			done = true
		}
		if start <= v.End && v.End < end {
			v.End = end
			done = true
		}

		if done == false && v.End < start {
			tr, _ := TimeRange.New()
			this.Ranges.InsertAfter(tr, e)
			done = true
		}

		if done {
			break
		}
	}

	if done == false {
		tr, _ := TimeRange.New()
		this.Ranges.PushFront(tr)
	}

	return nil
}

func (this *TimeRanges) Remove(start float64, end float64) error {
	return nil
}

func (this *TimeRanges) Len() int {
	return this.Ranges.Len()
}
