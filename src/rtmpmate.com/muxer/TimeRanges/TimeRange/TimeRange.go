package TimeRange

import ()

type TimeRange struct {
	Start float64
	End   float64
}

func New() (*TimeRange, error) {
	var tr TimeRange
	return &tr, nil
}
