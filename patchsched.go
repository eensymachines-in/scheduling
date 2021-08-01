package scheduling

type patchSchedule struct {
	*primarySched
}

func (pas *patchSchedule) ToTask() (Trigger, Trigger, int, int) {
	elapsed := ElapsedSecondsNow()
	var nr, fr Trigger
	// When its a patch schedule pre sleep is contextual as well.
	pre := pas.Delay()
	post := 0
	// Patch schedules are not circular
	// They allow pre sleep and are effective only between the triggers from top to bottom
	// so for all the cases the near trigger is the lower and the far one is the higher
	nr, fr = pas.lower, pas.higher
	if elapsed >= pas.lower.At() && elapsed < pas.higher.At() {
		// Case of in between ..no pre sleep
		post = pas.higher.At() - elapsed
	} else {
		// current time beyond the triggers, pre sleep comes into play
		post = pas.higher.At() - pas.lower.At()
		if elapsed < pas.lower.At() {
			pre += pas.lower.At() - elapsed
		}
		if elapsed >= pas.higher.At() {
			pre += 86400 - elapsed + pas.lower.At()
		}
	}
	return nr, fr, pre, post
}

// Please be ware here another cannot be a primary schedule
func (pas *patchSchedule) ConflictsWith(another Schedule) bool {
	// https://eensymachines-in.github.io/luminapi/schedule-conflicts
	// Read here patch schedules conflict with other patch schedules only in the case of overlap and intersection
	// in all other cases if the schedules are delayed incase of intersection
	outside, inside, overlap, coinc := overlapsWith(pas, another)
	// Getting if there's an intersection on the relays
	anLw, _ := another.Triggers()
	intersects := pas.lower.Intersects(anLw, false)
	if inside || outside || coinc {
		// Delay is added only if schedule intersects
		if intersects {
			// Delay is added to that schedule that comes later in the day
			if another.Midpoint() > pas.Midpoint() {
				another.AddDelay(pas.Delay())
			} else {
				pas.AddDelay(another.Delay())
			}
		}
		return false
	}
	if overlap && intersects {
		_, ok := another.(*patchSchedule)
		return ok // if the schedule type is not patchSchedule conflict cannot be determined
	}
	return false // if the schedules either does not overlap or does not intersect
}
