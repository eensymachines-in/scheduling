package scheduling

type patchSchedule struct {
	*primarySched
}

func (pas *patchSchedule) NearFarTrigger(elapsed int) (Trigger, Trigger, int, int) {
	var nr, fr Trigger
	// When its a patch schedule pre sleep is contextual as well.
	pre := 0
	post := 0
	// Patch schedules are not circular
	// They allow pre sleep and are effective only between the triggers from top to bottom
	// so for all the cases the near trigger is the lower and the far one is the higher
	nr, fr = pas.lower, pas.higher
	if elapsed >= pas.lower.At() && elapsed < pas.higher.At() {
		// Case of in between ..no pre sleep
		pre = 0
		post = pas.higher.At() - elapsed
	} else {
		// current time beyond the triggers, pre sleep comes into play
		post = pas.higher.At() - pas.lower.At()
		if elapsed < pas.lower.At() {
			pre = pas.lower.At() - elapsed
		}
		if elapsed >= pas.higher.At() {
			pre = 86400 - elapsed + pas.lower.At()
		}
	}
	return nr, fr, pre, post // for primary schedules pre sleep is always 0
}

func (pas *patchSchedule) ConflictsWith(another Schedule) bool {
	// Here schedules with same time slot (subset, overlaps, coincide) cannot have the same relays
	// if they have disjoint relays to work on.. then all of the above is allowed
	// for patch schedule, it cannot overlap / coincide with primary schedule
	if pas.overlapsWith(another) {
		_, ok := another.(*patchSchedule)
		if ok {
			// patch schedule when being assesed with othe patch schedule ..
			// 2 patch schedules can have overlaps incase they are operating on different relays
			anLw, _ := another.Triggers()
			if pas.lower.Intersects(anLw, false) {
				// Overlaps and also intersects .. so conflict
				return true
			}
			// Overlaps but has no intersection
			return false
		}
		return true // if overlaps and it isnt another patch schedule then there is conflict
	}
	return false // if there isnt any overlap then outright there isnt any conflict
}
