#### Clock schedules :
------------

We need schedules that run on __24 hour cycles.__ Schedules that repeat themselves over and over once set for any day. A user would like to define the behaviour  of any logic of actuation, and that logic needs to loop over the `clock`. 

Examples 

- Lights turned ON between 18:30-06:30 and OFF from 06:30-18:30
- Aquarium filter runs 6 hours a day between 13:00-19:00 and OFF throughout the rest of the day.

At any given point in the day the device wakes-up/boots it should __identify based on the current time, of what state it has to be in__ and then set up the cyclic sleep and change of states. 

For the light switching example - Lets assume the device wakes up / boots at 16:30, it should then turn the lights OFF, sleep for 2 hours and then flip the state. and sleep again till 06:30 _next morning_. The cyclic nature of the schedule allows us to let the algorithm to be agnostic of the date.

_Schedules such as these are driven by the seconds elapsed ahead of the midnight, irrespective of the date/day._

#### Seconds since midnight :
-------------

A day comprises of `86400` seconds, so any point in the day lineraly can be represented using elapsed since midnight. It is much convinient to define sleep times and also calculate overlaps for time ranges if time is represented using the same phenomenon. Though for user-level representation its much legible to keep it human readable string format. 

This module include functions that let you interconvert the 2 formats of time. 

#### 2 types of schedules :
-------------
Business logic needs us to define 2 types schedules. Not one type of schedule can suffice the wide range of needs. 
Here is an example from domain.

> A residential society has its GBM that has decided they need lights ON 12 hours a day and nothing more than that.- say 18:30-06:30. While respecting the directions of the GBM, there are a few floors who would need the lights to be on from 17:30, an extra hour. Which is understandable when there are senior citizens, who would need that extra ON time.

As you can notice from the above requirement, the exceptional ON  time is not cyclic. Beyond 17:30-18:30 it has no effect on the state of the lights. While the wider directive which affects 12 hours, has a implication beyond 18:30 as well, or before 06:30 

##### Primary schedules / cyclic schedules :
-----------

They are cyclic and imply the state over 24 hour clock. `06:30(OFF)-18:30(ON)` implies 18:30 to 06:30 the lights are ON, while ofcourse maintaining  that between those 12 hours, the lights are OFF. Lets for a moment think the device wakes up at 20:30, it will then `turn ON the lights` and sleep till 06:30 the next morning. Kindly read that carefully since there lies the subtle difference between primary and secondary schedules.

##### Secondary schedules / patch schedules :
-----------

They are seen more like exceptions to the above primary schedules, where beyond their said bracket they do not change the state. 
Lets assume a device wakes up / boots up at 20:30, considering the above case, the exception of 17:30-18:30 is not applicable here, so the device sees the time between 20:30-17:30(next day) as the sleep time. So unless the device finds itself in the middle of that exception time range it would not take any effect. 


#### JSON Schedules :
--------------

It all starts here, basic schedules are picked up from __json files__  a `schedules` array attribute having objects in the below format are expected. `JSONRelayState` is the one schedule. It has ids of the relay that would be signalled, ON/OFF times as string, and `primary` bool attribute that indicates 
the type of schedule. If not primary the schedule is regarded as `patch`

```go
type JSONRelayState struct {
	ON      string   `json:"on"`
	OFF     string   `json:"off"`
	IDs     []string `json:"ids"`
	Primary bool     `json:"primary"`
}
```
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