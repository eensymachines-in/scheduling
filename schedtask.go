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
