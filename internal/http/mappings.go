package http

import (
	"rockets/internal/http/gen"
	"rockets/internal/rocket"
)

// stateToServer converts a rocket.State to a gen.RocketState.
func stateToServer(state rocket.State) gen.RocketState {
	var status gen.RocketStateStatus
	switch state.Status {
	case rocket.StatusLaunched:
		status = gen.LAUNCHED
	case rocket.StatusExploded:
		status = gen.EXPLODED
	}

	return gen.RocketState{
		CurrentSpeed:               state.CurrentSpeed,
		Id:                         state.ID,
		LastProcessedMessageNumber: state.LastProcessedMessageNumber,
		LastUpdateTime:             state.LastUpdateTime,
		Mission:                    state.Mission,
		Reason:                     state.Reason,
		Status:                     status,
		Type:                       state.Type,
	}
}
