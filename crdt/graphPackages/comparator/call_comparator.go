package comparator

import "myproject/operation"

type CallComparator struct{}

func (comp CallComparator) String() string {
	return "CallComparator: a.Time.Compare(o.Time)"
}

func (comp CallComparator) Compare(a, o *operation.Call) (int, error) {
	return a.Time.Compare(o.Time)
}