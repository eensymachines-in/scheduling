package scheduling

/*All the public facing interfaces here
Using this package from client code should be easy if you get your head around this file
*/

import (
	"fmt"
)

// Schedule : is the handle for external packages
type Schedule interface {
	Triggers() (Trigger, Trigger)
	Duration() int
	ToTask(elapsed int) ScheduleTask
	ConflictsWith(another Schedule) bool
	Midpoint() int
	Delay() int
	AddDelay(prior int) Schedule
	Conflicts() int
	AddConflict() Schedule
	Close()
}

// ScheduleTask : when the schedule calculated in context of current time it boils down to pre sleep > trigger > post sleep > trigger > end task
type ScheduleTask interface {
	Apply(ok, cancel chan interface{}, send chan []byte, err chan error)
	Loop(cancel, interrupt chan interface{}, send chan []byte, loopErr chan error)
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

// NewSchedule : makes a new TriggeredSchedul, will take 2 triggers
func NewSchedule(trg1, trg2 Trigger, primary bool) (Schedule, error) {
	l, h, err := sortTriggers(trg1, trg2)
	if err != nil {
		return nil, err
	}
	if !trg1.Intersects(trg2, true) || trg1.Coincides(trg2) {
		// When triggers are paired in a schedule they have to be intersecting and not coincident
		return nil, fmt.Errorf("%s-%s Triggers for the schedule are either not exactly intersecting or are coinciding", trg1, trg2)
	}
	if primary {
		return &primarySched{l, h, 0, 0}, nil
	}
	return &patchSchedule{&primarySched{l, h, 0, 0}}, nil

}

// JSONRelayState : relaystate but in json format
// ================================== Json Relay state is for file reads ============================
// Making a relay state from a json file
type JSONRelayState struct {
	ON      string   `json:"on"`
	OFF     string   `json:"off"`
	IDs     []string `json:"ids"`
	Primary bool     `json:"primary"`
}

// ToSchedule : reads from json and pumps up a schedule
// this saves you the trouble of making a schedule via code,
// from a json file it can read up a relaystate and convert that to schedule
func (jrs *JSONRelayState) ToSchedule() (Schedule, error) {
	onTm, err := TimeStr(jrs.ON).ToElapsedTm()
	if err != nil {
		return nil, fmt.Errorf("Failed to read ON time for schedule")
	}
	offTm, err := TimeStr(jrs.OFF).ToElapsedTm()
	if err != nil {
		return nil, fmt.Errorf("Failed to read OFF time for schedule")
	}
	offs := []*RelayState{}
	ons := []*RelayState{}
	for _, id := range jrs.IDs {
		offs = append(offs, &RelayState{byte(0), id})
		ons = append(ons, &RelayState{byte(1), id})
	}
	trg1, trg2 := NewTrg(offTm, offs...), NewTrg(onTm, ons...)
	return NewSchedule(trg1, trg2, jrs.Primary)

}

// NewScheduleTask Makes a simple new schedule task
func NewScheduleTask(nr, fr Trigger, pre, post int) ScheduleTask {
	return &schedTask{nr, fr, pre, post}
}
