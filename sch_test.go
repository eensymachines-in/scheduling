package scheduling

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func TestReadSchedules(t *testing.T) {
	// TestReadSchedules : just about reading schedules and converting them to schedule objects
	scheds, err := ReadScheduleFile("test_sched.json")
	if err != nil {
		t.Error(err)
		panic("TestReadSchedules: error reading schedule files")
	}
	for _, s := range scheds {
		t.Logf("%s\n", s)
	}
}

func TestScheduleConflicts(t *testing.T) {
	scheds, err := ReadScheduleFile("test_sched3.json")
	if err != nil {
		t.Error(err)
		panic("TestReadSchedules: error reading schedule files")
	}
	if len(scheds) == 0 {
		t.Error("")
	}
	primaSched := scheds[0]
	for _, s := range scheds[1:] {
		if primaSched.ConflictsWith(s) {
			t.Logf("%s has conflict with primary schedule", s)
		}
	}
}
func listenOnSend(t *testing.T, errx chan error, send chan []byte, cancel chan interface{}) {
	for {
		select {
		case <-time.After(1 * time.Second):
			continue
		case err := <-errx:
			t.Errorf("Error from shcedule application %s", err)
		case msg := <-send:
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", "localhost", "35001"))
			if err != nil {
				t.Errorf("Failed to connect to socket over TCP %s", err)
			}
			_, err = conn.Write(msg)
			if err != nil {
				t.Errorf("Error writing message to TCP sock %s", err)
			}
			conn.Close()
		case <-cancel:
			return
		}
	}
}
func TestScheduleApply(t *testing.T) {
	scheds, err := ReadScheduleFile("test_sched3.json")
	if err != nil {
		t.Error(err)
		panic("TestReadSchedules: error reading schedule files")
	}
	if len(scheds) == 0 {
		t.Error("")
	}
	// Send channel here is nil since we just want to log the tcp messages
	// we may write more code to get the messages across TCP
	ctx := &SchedCtx{Ok: make(chan interface{}, 1), Cancel: make(chan interface{}), Send: make(chan []byte, 10), Err: make(chan error, 10)}
	defer ctx.Close()
	for _, s := range scheds {
		if s.Conflicts() == 0 {
			go Apply(s, ctx)
		} else {
			t.Logf("%s has %d conflicts \n", s, s.Conflicts())
		}
	}
	go listenOnSend(t, ctx.Err, ctx.Send, ctx.Cancel)
	<-time.After(60 * time.Minute)
	t.Log("Now closing the context..")
	// this closes the schedule context and hence all the tasks
}

func TestScheduleLoop(t *testing.T) {
	scheds, err := ReadScheduleFile("test_sched3.json")
	if err != nil {
		t.Error(err)
		panic("TestReadSchedules: error reading schedule files")
	}
	if len(scheds) == 0 {
		t.Error("")
	}
	sigctx := SignalCtx{Cancel: make(chan interface{})}
	loopctx := SchedLoopCtx{Interrupt: make(chan interface{})}
	send := make(chan []byte, 10)
	errx := make(chan error, 10)
	for _, s := range scheds {
		if s.Conflicts() == 0 {
			go Loop(s, &sigctx, &loopctx, send, errx)
		} else {
			t.Logf("%s has %d conflicts \n", s, s.Conflicts())
		}
	}
	go listenOnSend(t, errx, send, sigctx.Cancel)
	<-time.After(60 * time.Minute)
	t.Log("Now closing the context..")
	loopctx.Close()
	sigctx.Close()
	// this closes the schedule context and hence all the tasks
}
