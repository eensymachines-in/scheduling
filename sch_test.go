package scheduling

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
func TestWriteSchedules(t *testing.T) {
	sojrs := SliceOfJSONRelayState{
		{ON: "06:30 PM", OFF: "06:30 AM", IDs: []string{"IN1", "IN2"}, Primary: true},
		{ON: "04:30 PM", OFF: "06:13 PM", IDs: []string{"IN1"}, Primary: false},
	}
	err := WriteScheduleFile("test_sched.json", sojrs)
	if err != nil {
		panic(err)
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
	stop := make(chan interface{})
	send := make(chan []byte)
	errx := make(chan error)
	defer close(stop)
	defer close(errx)
	defer close(send)
	// Send channel here is nil since we just want to log the tcp messages
	// we may write more code to get the messages across TCP
	for _, s := range scheds {
		if s.Conflicts() == 0 {
			call, _ := Apply(s, stop, send, errx)
			go call()
		} else {
			t.Logf("%s has %d conflicts \n", s, s.Conflicts())
		}
	}
	go listenOnSend(t, errx, send, stop)
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
	stop := make(chan interface{})
	interrupt := make(chan interface{})
	send := make(chan []byte)
	errx := make(chan error)
	defer close(stop)
	defer close(interrupt)
	defer close(errx)
	defer close(send)
	for _, s := range scheds {
		if s.Conflicts() == 0 {
			go Loop(s, stop, interrupt, send, errx)
		} else {
			t.Logf("%s has %d conflicts \n", s, s.Conflicts())
		}
	}
	go listenOnSend(t, errx, send, stop)
	<-time.After(9999 * time.Second)
	t.Log("Now closing the test..")
}

func TestOverlappingSchedules(t *testing.T) {
	scheds, err := ReadScheduleFile("test_sched4.json")
	if err != nil {
		t.Error(err)
		panic("TestReadSchedules: error reading schedule files")
	}
	for i, s := range scheds {
		for _, ss := range scheds[i+1:] {
			ou, in, ov, co := overlapsWith(s, ss)
			t.Logf("%s/%d-%s/%d", s, s.Delay(), ss, ss.Delay())
			t.Logf("outside: %t, inside: %t, overlap: %t, coincide: %t", ou, in, ov, co)
			t.Logf("Conflicts: %t", s.ConflictsWith(ss))
		}
	}
}

// TestToSchedules : will test conversion of slice of json relay state to schedules
func TestToSchedules(t *testing.T) {
	jrs := SliceOfJSONRelayState{
		{ON: "06:30 PM", OFF: "06:30 AM", IDs: []string{"IN1", "IN2"}, Primary: true},
		{ON: "04:30 PM", OFF: "06:29 PM", IDs: []string{"IN1"}, Primary: false},
	}
	scheds := []Schedule{}
	err := jrs.ToSchedules(&scheds)
	assert.Nil(t, err, "Error converting to a slice of Schedules")
	if err != nil {
		return
	}
	for _, s := range scheds {
		t.Log(s)
		t.Log(s.Conflicts())
	}
	scheds, err = ReadScheduleFile("./test_sched2.json")
	assert.Nil(t, err, "Unexpected error reading schedules from file")
	if err != nil {
		return
	}
	for _, s := range scheds {
		t.Logf("%v:%d", s, s.Conflicts())
	}
}
