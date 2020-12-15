package scheduling

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/mgutz/logxi/v1"
)

type schedTask struct {
	Nr   Trigger
	Fr   Trigger
	Pre  int
	Post int
}

func (st *schedTask) Apply(ok, cancel chan interface{}, send chan []byte, err chan error) {
	// If on delay, this schedule has preceedence
	if st.Pre > 0 {
		fmt.Printf("Pre sleep %d\n", st.Pre)
		// this will work as expected even when pre=0, but the problem is it sill still allow the processor to jump to the next task
		<-time.After(time.Duration(st.Pre) * time.Duration(1*time.Second))
	}
	byt, e := json.Marshal(st.Nr)
	if e != nil { // state of the trigger is applied
		err <- fmt.Errorf("Schedule/Apply: Failed to marshall trigger data - %s", e)
		return
	}
	send <- byt
	select {
	// sleep duration is always a second extra than the sleep time
	// so that incase the processor is fast enough this will still be in the next slot
	case <-time.After(time.Duration(st.Post+1) * time.Duration(1*time.Second)):
		log.Info("End of task.. ")
		if byt, e = json.Marshal(st.Fr); e != nil {
			err <- fmt.Errorf("Schedule/Apply: Failed to marshall trigger data - %s", e)
			return
		}
		send <- byt
		ok <- struct{}{}
	case <-cancel:
		log.Warn("Task/Apply: Interruption")
	}
	return
}

// Loop : this shall loop the schedule forever till there is a interruption or the schedule application fails
func (st *schedTask) Loop(cancel, interrupt chan interface{}, send chan []byte, loopErr chan error) {
	// this channnel communicates the ok from apply function
	// The loop still does not indicate done unless ofcourse the done <-nil
	ok := make(chan interface{}, 1)
	defer close(ok)
	stop := make(chan interface{}) // this is to stop the currently running schedule
	for {
		st.Apply(ok, stop, send, loopErr) // applies the schedul infinitely
		select {
		case <-cancel:
		case <-interrupt:
			close(stop)
			log.Warn("Running schedule is stopped or interrupted, now closing the loop as well")
			return
		case <-ok:
			// this is when the schedule has done applying for one cycle
			// will go back to applying the next schedule for the then current time
		}
	}
}
