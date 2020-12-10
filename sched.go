package scheduling

import (
	"fmt"
)

// Schedule : is the handle for external packages
type Schedule interface {
	Triggers() (Trigger, Trigger)
	Duration() int
	NearFarTrigger(elapsed int) (Trigger, Trigger, int, int)
	ConflictsWith(another Schedule) bool
	Midpoint() int
	Close()
	Apply(ok, cancel chan interface{}, send chan []byte, err chan error)
}

func sortTriggers(trg1, trg2 Trigger) (l, h Trigger, e error) {
	if trg1.At() == trg2.At() || trg1 == nil || trg2 == nil {
		e = fmt.Errorf("ERROR/sortTriggers: triggers cannot be overlapping, or nil")
		return
	}
	if trg1.At() < trg2.At() {
		l, h = trg1, trg2
	} else {
		l, h = trg2, trg1
	}
	return
}

// NewPrimarySchedule : makes a new TriggeredSchedul, will take 2 triggers
func NewPrimarySchedule(trg1, trg2 Trigger) (Schedule, error) {
	l, h, err := sortTriggers(trg1, trg2)
	if err != nil {
		return nil, err
	}
	return &primarySched{l, h}, nil
}

// NewPatchSchedule : this makes a new patch schedule object
func NewPatchSchedule(trg1, trg2 Trigger) (Schedule, error) {
	l, h, err := sortTriggers(trg1, trg2)
	if err != nil {
		return nil, err
	}
	return &patchSchedule{&primarySched{l, h}}, nil
}
