package scheduling

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// primarySched : this schedule is the longer schedule and in all the cases there is only one of this
// primarySched is circular and beyond the triggers applies the last valid state of the trigger
type primarySched struct {
	lower  Trigger
	higher Trigger
	// whenever the schedule gets in a conflict the LHS induces increment in the RHS conflict
	conflicts int
	delay     int // increasing this will increment the preceedence since this will be applied after a delay
}

func (ps *primarySched) Conflicts() int {
	return ps.conflicts
}
func (ps *primarySched) AddConflict() Schedule {
	ps.conflicts++
	return ps
}
func (ps *primarySched) Delay() int {
	return ps.delay
}
func (ps *primarySched) AddDelay(prior int) Schedule {
	ps.delay = prior
	ps.delay++
	return ps
}
func (ps *primarySched) Triggers() (Trigger, Trigger) {
	return ps.lower, ps.higher
}
func (ps *primarySched) Duration() int {
	return ps.higher.At() - ps.lower.At()
}
func (ps *primarySched) Midpoint() int {
	return (ps.Duration() / 2) + ps.lower.At()
}
func (ps *primarySched) Close() {
	// For now all what the schedule does when closing is just ouput a log message
	log.Infof("%s Schedule is now closing", ps)
}
func (ps *primarySched) String() string {
	return fmt.Sprintf("%s - %s %v %v", tmStrFromUnixSecs(ps.lower.At()), tmStrFromUnixSecs(ps.higher.At()), ps.lower.RelayIDs(), ps.higher.RelayIDs())
}

// NearFarTrigger : in context of the current time, this helps to get the triggers that are near or far
// For any schedule when its applied - pre sleep - nr state apply - post sleep - fr state apply
// For a primary schedule its thought to be circular, meaning to say : if beyond the trigger bounds the higher trigger is applied
func (ps *primarySched) ToTask(elapsed int) ScheduleTask {
	// for primary schedule nr trigger will be applied then, sleep, then fr state
	// for primary schedule there is no pre sleep - since its circular and applies beyond the 2 triggers as well
	var nr, fr Trigger
	var post int
	pre := ps.Delay()
	if elapsed >= ps.lower.At() && elapsed < ps.higher.At() {
		nr, fr = ps.lower, ps.higher
		post = ps.higher.At() - elapsed
	} else {
		nr, fr = ps.higher, ps.lower
		if elapsed < ps.lower.At() {
			post = ps.lower.At() - elapsed
		}
		if elapsed > ps.higher.At() {
			post = 86400 - elapsed + ps.lower.At()
		}
	}
	return NewScheduleTask(nr, fr, pre, post)
}

func (ps *primarySched) overlapsWith(another Schedule) bool {
	// Midpoints are distance of the half time since midnight for any schedule
	mdpt1, mdpt2 := ps.Midpoint(), another.Midpoint()
	// half duration of each schedule
	hfdur1, hfdur2 := ps.Duration()/2, another.Duration()/2
	// getting the absolute of the midpoint distance
	mdptdis := mdpt1 - mdpt2
	if mdptdis < 0 {
		mdptdis = -mdptdis
	}
	// Getting the larger of the 2 schedules
	var min, max int
	if hfdur1 <= hfdur2 {
		min, max = hfdur1, hfdur2
	} else {
		min, max = hfdur2, hfdur1
	}
	if outside, inside := (mdptdis > (hfdur1 + hfdur2)), ((mdptdis + min) < max); outside || inside {
		// case when the schedules are clearing and not interferring with one another
		// either one schedule is inside the other or on one side
		if inside {
			// Here we would want to adjust the preceedence too.
			another.AddDelay(ps.Delay())
		}
		return false
	}
	// all other cases the schedules are either partially/exactly overlapping
	return true
}

// ConflictsWith : checks to see partial overlapping of schedules
func (ps *primarySched) ConflictsWith(another Schedule) bool {
	_, ok := another.(*primarySched)
	if ok {
		// Always conflicts with other primary schedule
		// overlaps are checked for circular and non-cicrular schedule
		return true
	}
	return ps.overlapsWith(another)
}

// // Loop : this shall loop the schedule forever till there is a interruption or the schedule application fails
// func (ps *primarySched) Loop(cancel, interrupt chan interface{}, send chan []byte, loopErr chan error) {
// 	// this channnel communicates the ok from apply function
// 	// The loop still does not indicate done unless ofcourse the done <-nil
// 	ok := make(chan interface{}, 1)
// 	defer close(ok)
// 	stop := make(chan interface{}) // this is to stop the currently running schedule
// 	for {
// 		ps.Apply(ok, stop, send, loopErr) // applies the schedul infinitely
// 		select {
// 		case <-cancel:
// 		case <-interrupt:
// 			close(stop)
// 			log.Warn("Running schedule is stopped or interrupted, now closing the loop as well")
// 			return
// 		case <-ok:
// 			// this is when the schedule has done applying for one cycle
// 			// will go back to applying the next schedule for the then current time
// 		}
// 	}
// }
