package heartbeat

import "dhs/internal/model"

type Transition struct {
	From    model.Status
	To      model.Status
	Reason  string
	Trigger string
}

func ValidateTransition(t Transition) error {
	if t.From == t.To {
		return nil
	}
	if !CanTransition(t.From, t.To) {
		return ErrIllegalTransition{From: t.From, To: t.To}
	}
	if t.Reason == "" || t.Trigger == "" {
		return ErrMissingTransitionMetadata{}
	}
	return nil
}

type ErrIllegalTransition struct{ From, To model.Status }

func (e ErrIllegalTransition) Error() string {
	return "illegal transition: " + string(e.From) + " -> " + string(e.To)
}

type ErrMissingTransitionMetadata struct{}

func (ErrMissingTransitionMetadata) Error() string {
	return "transition reason and trigger are required"
}
