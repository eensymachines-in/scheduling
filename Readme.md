#### Clock schedules :
------------

Clockwork automation has many applications in IoT products. Typically a user would want to set a clock to certain tasks and expect the tasks to repeat indefinetly. - or atleast until the user can intervene and change the clock configuration. We need schedules that run on __24 hour cycles.__ Schedules that repeat themselves in a loop, once set for any day. User shall be able to change such a _clock_ and such the new clock is in effect the device follow the same.

##### Examples

- Domestic lights turned ON between `18:30-06:30` and OFF from `06:30-18:30`
- Aquarium filter runs 6 hours a day between `13:00-19:00` and OFF throughout the rest of the day.

At any given point in the day the device wakes-up/boots it should __identify based on the current time, of what state it has to be in__ and then set up the cyclic sleep and change of states. An example of domestic illumination control - Lets assume the device wakes up / boots at 16:30, it should then turn the lights OFF, sleep for 2 hours (Since the lights are to be switched on at 18:30 for 12 hours ahead.) and then flip the state. and sleep again till 06:30 _next morning_. The cyclic nature of the schedule allows us to let the algorithm to be agnostic of the date and day. 

> Schedules such as these are driven by the seconds elapsed ahead of the midnight, irrespective of the date/day

#### Seconds since midnight :
-------------

A day comprises of `86400` seconds, so any point in the day lineraly can be represented using elapsed since midnight. It is much convinient to define sleep times and also calculate overlaps for time ranges if time is represented using the same phenomenon. Though for user-level representation its much legible to keep it human readable string format. 

__This module include functions that let you interconvert the 2 formats of time.__ 

#### 2 types of schedules :
-------------
Business logic needs us to define 2 types schedules. Not one type of schedule can suffice the wide range of needs. 
Here is an example from domain.

> A residential society has its GBM that has decided they need lights ON 12 hours a day and nothing more than that.- say 18:30-06:30. While respecting the directions of the GBM, there are a few floors who would need the lights to be on from 17:30, an extra hour. Which is understandable when there are senior citizens, who would need that extra ON time.

As you can notice from the above requirement, the exceptional ON  time is not cyclic. `Beyond 17:30-18:30` _it has no effect on the state of the lights._ While the wider directive which affects 12 hours, has a implication beyond 18:30 as well, or before 06:30.

Schedules define a slot in which a state of the lights is explicit, while anything outside the slot is considered reverse implicitly. Exceptions on the other hand though are confined to their effect within the defined slot. Exceptions are also schedules, just that they operate _typically_.

##### Primary schedules / cyclic schedules :
-----------

`They are cyclic` and imply the state over 24 hour clock. `06:30(OFF)-18:30(ON)` implies 18:30 to 06:30 the lights are ON, while ofcourse maintaining that between those 12 hours, the lights are OFF. Lets for a moment think the device wakes up at 20:30, it will then `turn ON the lights` and sleep till 06:30 the next morning. 
_Kindly read that carefully since there lies the subtle difference between primary and secondary schedules._

##### Secondary schedules / patch schedules :
-----------

They are seen more like exceptions/patches to the above primary schedules, where beyond their said bracket they do not change the state. `They aren't cyclic`
Lets assume a device wakes up / boots up at 20:30, considering the above case, the exception of 17:30-18:30 is not applicable here, so the device sees the time between 20:30-17:30(next day) as the sleep time. __Hence unless the device finds itself in the middle of that exception time range it would not take any effect.__ 

#### JSON Schedules :
--------------

It all starts here, basic schedules are picked up from __json files__  a `schedules` array attribute having objects in the below format are expected. `JSONRelayState` is the one schedule. It has ids of the relay that would be signalled, ON/OFF times as string, and `primary` bool attribute that indicates 
the type of schedule. If not primary the schedule is regarded as `patch`

schedules in JSON look somewhat like this.
```json 
{
    "schedules": [
        {"on":"05:00 PM", "off":"12:00 PM","primary":true, "ids":["IN1","IN2","IN3","IN4"]},
        {"on":"04:45 PM", "off":"06:24 PM","primary":false, "ids":["IN4"]},
        {"on":"06:37 PM", "off":"06:35 PM","primary":false, "ids":["IN4"]}
    ]
}
```
```go
type JSONRelayState struct {
	ON      string   `json:"on"`
	OFF     string   `json:"off"`
	IDs     []string `json:"ids"`
	Primary bool     `json:"primary"`
}

func (jrs *JSONRelayState) ToSchedule() (Schedule, error)

```
`JSONRelayState` read-in from the json file can be converted to a schedule with a simple method. This can make the relay states correctly and pack them into 2 trigger schedule.
A schedule is nothing but a set of 2 triggers, one - ON other OFF each associated with relay pins. A single schedule can be applied to one or many relay pins at a time.

#### Reading JSON schedules in :
--------------

We then need a function that would pick such json schedules and check/mark them for conflicts they have.
Conflicting schedules are often neglected and only those with no conflicts are spawned / Looped

```go
scheds, e := scheduling.ReadScheduleFile("path/to/file.json")
if e != nil {
    log.Error(e)
    panic("Failed to read schedule file, check json and retry")
}
```

#### Starting schedules as routines:
---------

Once you have a slice of schedules, all what remains to start is the schedules in a loop. Check for conflicts, if no conflicts the `Loop` function can be used to spawn new schedules.

```go
for _, s := range scheds {
    if s.Conflicts() == 0 {
        go scheduling.Loop(s, cancel, interrupt, send, errx)
    } else {
        log.Warnf("%s has %d conflicts \n", s, s.Conflicts())
    }
}
```

#### Applying schedules (for one cycle):
---------

```go

func Apply(sch Schedule, stop chan interface{}, send chan []byte, errx chan error) (func(), chan interface{}) 

// sch  : Schedule object that needs to be applied
// stop : close this channel to indicate if the Apply function call needs to abort
// send : status of the application is communicated over this channel
// errx : any error applying the schedule will be on this channel

// func() : callback to start the application 
// ok chan interface{} : when this channel is closed it indicates the schedule has been successfuly applied once

stop := make(chan interface{})
defer close(stop)
send := make(chan []byte,10)
defer close(send)
errx := make(chan error,10)
defer close(errx)
call, ok := Apply(sch, stop, send, errx)
go call()
```

While schedules are meant to be applied in a loop till there is a interruption signal, they are designed to be applied for one complete cycle. Between 2 states of a relay there is a sleep routine. If the current time is beyond the schedule effect then, it demands an extra sleep routine as well.

Lets understand this from an example

- 06:30 AM to 06:30 PM Primary, current time is 17:30 : here state of 06:30AM is applied, schedule sleeps for 12 hours, and then applies state of 06:30PM
- 06:30 AM to 06:30 PM Primary, current time is 19:30 : here state of 06:30PM is applied, schedule sleeps till 06:30AM, and then applies state of 06:30AM
- 04:30PM to 05:30PM Patch, current time is 16:35, state at 04:30PM is applied and sleeps till 05:30 and then applies its state
- 04:30PM to 05:30PM Patch, current time is 17:35, No state is applied, sleeps till 04:30 next day, then applies its state, sleeps for 1 hour and then applies 05:30 state

There are 2 types of schedules and each behaves distinctly depending on where the current time is when applied. Primary schedules are considered to be cyclic while patch schedules are effective only within a time zone.

To Apply a schedule all what you need to do is pass the schedule to the `Apply` function and run it as a go-routine