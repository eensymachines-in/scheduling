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
	return fmt.Sprintf("%s - %s %v %v", TmStrFromUnixSecs(ps.lower.At()), TmStrFromUnixSecs(ps.higher.At()), ps.lower.RelayIDs(), ps.higher.RelayIDs())
}

// NearFarTrigger : in context of the current time, this helps to get the triggers that are near or far
// For any schedule when its applied - pre sleep - nr state apply - post sleep - fr state apply
// For a primary schedule its thought to be circular, meaning to say : if beyond the trigger bounds the higher trigger is applied
func (ps *primarySched) ToTask() (Trigger, Trigger, int, int) {
	// for primary schedule nr trigger will be applied then, sleep, then fr state
	// for primary schedule there is no pre sleep - since its circular and applies beyond the 2 triggers as well
	elapsed := ElapsedSecondsNow()
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
	return nr, fr, pre, post
}

// ConflictsWith : checks to see partial overlapping of schedules
func (ps *primarySched) ConflictsWith(another Schedule) bool {
	_, ok := another.(*primarySched)
	if ok {
		// Always conflicts with other primary schedule
		// overlaps are checked for circular and non-cicrular schedule
		return true
	}
	outside, inside, overlap := overlapsWith(ps, another)
	if outside || inside {
		// When its the primary schedule the patch schedule will always be given proceedence
		another.AddDelay(ps.Delay())
	}
	return overlap
}
