package heartbeat

import "dhs/internal/model"

func CanTransition(from, to model.Status) bool {
	if from == to {
		return true
	}
	switch from {
	case model.Registered:
		return to == model.Online
	case model.Online:
		return to == model.Lost
	case model.Lost:
		return to == model.Recovering || to == model.Offline
	case model.Recovering:
		return to == model.Online || to == model.Lost || to == model.Offline
	case model.Offline:
		return to == model.Registered
	}
	return false
}
