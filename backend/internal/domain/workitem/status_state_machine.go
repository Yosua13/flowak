package workitem

// CanTransitionStatus validates if transitioning from status 'from' to 'to' is permitted by the Kanban state machine.
func CanTransitionStatus(from, to Status) bool {
	if from == to {
		return true
	}
	allowed := map[Status]map[Status]bool{
		StatusBacklog: {
			StatusReady:    true,
			StatusCanceled: true,
		},
		StatusReady: {
			StatusInProgress: true,
			StatusBlocked:    true,
			StatusCanceled:   true,
		},
		StatusInProgress: {
			StatusInReview: true,
			StatusBlocked:  true,
			StatusReady:    true,
			StatusCanceled: true,
		},
		StatusInReview: {
			StatusDone:       true,
			StatusInProgress: true,
			StatusBlocked:    true,
			StatusCanceled:   true,
		},
		StatusBlocked: {
			StatusReady:      true,
			StatusInProgress: true,
			StatusCanceled:   true,
		},
		StatusDone:     {},
		StatusCanceled: {},
	}
	return allowed[from][to]
}
