package scheduling

/*All the public facing interfaces here
Using this package from client code should be easy if you get your head around this file
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Schedule : is the handle for external packages
type Schedule interface {
	Triggers() (Trigger, Trigger)
	Duration() int
	ConflictsWith(another Schedule) bool
	Midpoint() int
	Delay() int
	AddDelay(prior int) Schedule
	Conflicts() int
	AddConflict() Schedule
	Close()
	ToTask() (Trigger, Trigger, int, int)
}

// SignalCtx : since signals communicate on cancellation only
type SignalCtx struct {
	Cancel chan interface{}
	once   sync.Once
}

// Close :Closes down the channels within
func (sc *SignalCtx) Close() {
	sc.once.Do(func() {
		close(sc.Cancel)
	})
}

// SchedCtx : for async routines on schedules this is helpful as a flyweight object
type SchedCtx struct {
	Ok     chan interface{}
	Cancel chan interface{}
	Send   chan []byte
	Err    chan error
	once   sync.Once
}

// Close :Closes down the channels within
func (schc *SchedCtx) Close() {
	// safe closing all the channels only once
	// https://go101.org/article/channel-closing.html
	schc.once.Do(func() {
		if schc.Cancel != nil {
			close(schc.Cancel)
		}
		if schc.Ok != nil {
			close(schc.Ok)
		}
		if schc.Send != nil {
			close(schc.Send)
		}
		if schc.Err != nil {
			close(schc.Err)
		}
	})
}

// SchedLoopCtx : loops are often interrupted but not closed, this often indicates restarting the loop
type SchedLoopCtx struct {
	Interrupt chan interface{}
	once      sync.Once
}

// Close :Closes down the channels within
func (slc *SchedLoopCtx) Close() {
	slc.once.Do(func() {
		close(slc.Interrupt)
	})
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

// Apply : applies the schedule once for a cycle pre>state>post>state
func Apply(sch Schedule, stop chan interface{}, send chan []byte, errx chan error) (func(), chan interface{}) {
	ok := make(chan interface{}, 1)
	return func() {
		select {
		case <-ok:
			ok = make(chan interface{}, 1)
			defer close(ok)
		default:
			// channel is not closed
			defer close(ok)
		}
		if sch == nil {
			errx <- fmt.Errorf("Schedule/Apply: Null schedule, cannot apply")
			return
		}
		nr, fr, pre, post := sch.ToTask()
		log.Debugf("Near: %s Far: %s Pre: %d Post: %d\n", nr, fr, pre, post)
		if pre > 0 {
			// this will work as expected even when pre=0, but the problem is it sill still allow the processor to jump to the next task
			<-time.After(time.Duration(pre) * time.Duration(1*time.Second))
		}
		byt, e := json.Marshal(nr)
		if e != nil { // state of the trigger is applied
			errx <- fmt.Errorf("Schedule/Apply: Failed to marshall trigger data - %s", e)
			return
		}
		if send != nil {
			send <- byt
		} else {
			log.Debugf("TCP: %s", string(byt))
		}

		select {
		// sleep duration is always a second extra than the sleep time
		// so that incase the processor is fast enough this will still be in the next slot
		case <-time.After(time.Duration(post+1) * time.Duration(1*time.Second)):
			log.Info("End of post duration")
			if byt, e = json.Marshal(fr); e != nil {
				errx <- fmt.Errorf("Schedule/Apply: Failed to marshall trigger data - %s", e)
				return
			}
			if send != nil {
				send <- byt
			} else {
				log.Debugf("TCP: %s", string(byt))
			}
			ok <- struct{}{}
		case <-stop:
			log.Warn("Task/Apply: Interruption\n")
		}
	}, ok
}

// Loop : this shall apply the schedule infinetly till the schedule is running fine
func Loop(sch Schedule, cancel, interrupt chan interface{}, send chan []byte, errx chan error) {
	stop := make(chan interface{})
	defer close(stop)
	for {
		call, ok := Apply(sch, stop, send, errx)
		go call()
		select {
		case <-cancel:
		case <-interrupt:
			// Incase there's a signal interruption or from file change, the loop will have to quit its infinite nature
			return
		case <-ok:
			// this is when the schedule has done applying for one cycle
			// will go back to applying the next schedule for the then current time
		}
	}
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

// ReadScheduleFile : just so that we can read json schedule file, and get slice of schedules
// we have also added some conflict detection in here
// Call this from the client function to get schedules with their conflict numbers
func ReadScheduleFile(file string) ([]Schedule, error) {
	jsonFile, _ := os.Open(file)
	// Reading bytes from the file and unmarshalling the same to struct values
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	jsonFile.Close() // since this returns a closure, the call to this cannot be deferred
	type conf struct {
		Schedules []JSONRelayState `json:"schedules"`
	}
	c := conf{}
	json.Unmarshal(bytes, &c)
	scheds := []Schedule{}
	// converting from json schedules to schedule object slice
	for _, s := range c.Schedules {
		sched, err := s.ToSchedule()
		if err != nil {
			return nil, err
		}
		scheds = append(scheds, sched)
	}
	// flagging conflicts
	for i, s := range scheds {
		for _, ss := range scheds[i+1:] {
			if s.ConflictsWith(ss) {
				ss.AddConflict()
			}
		}
	}
	return scheds, nil
}
