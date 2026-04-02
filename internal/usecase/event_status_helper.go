package usecase

import "github.com/Back-Seat-Boy/kumpul-be/internal/model"

func ensureEventNotCancelled(event *model.Event) error {
	if event != nil && event.Status == model.EventStatusCancelled {
		return model.ErrEventCancelled
	}
	return nil
}

func isValidEventStatusTransition(current, target model.EventStatus) bool {
	if current == target {
		return true
	}
	if current == model.EventStatusCompleted {
		return false
	}
	if target == model.EventStatusCancelled {
		return true
	}
	switch current {
	case model.EventStatusVoting:
		return target == model.EventStatusConfirmed
	case model.EventStatusConfirmed:
		return target == model.EventStatusOpen
	case model.EventStatusOpen:
		return target == model.EventStatusPaymentOpen
	case model.EventStatusPaymentOpen:
		return target == model.EventStatusCompleted
	default:
		return false
	}
}
