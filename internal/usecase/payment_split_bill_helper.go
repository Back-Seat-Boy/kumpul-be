package usecase

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
)

type splitBillComputation struct {
	SubtotalByParticipant map[string]int
	TaxByParticipant      map[string]int
	TotalByParticipant    map[string]int
	ItemDetails           []model.SplitBillItemDetail
	ItemsSubtotal         int
	GrandTotal            int
}

func buildSplitBillModels(paymentID string, items []model.SplitBillItemInput) []*model.SplitBillItem {
	result := make([]*model.SplitBillItem, 0, len(items))
	for _, item := range items {
		billItem := &model.SplitBillItem{
			ID:        uuid.New().String(),
			PaymentID: paymentID,
			Name:      strings.TrimSpace(item.Name),
			Price:     item.Price,
		}
		for _, participantID := range item.ParticipantIDs {
			billItem.Assignments = append(billItem.Assignments, model.SplitBillItemAssignment{
				ID:            uuid.New().String(),
				ItemID:        billItem.ID,
				ParticipantID: participantID,
			})
		}
		result = append(result, billItem)
	}
	return result
}

func validateSplitBillInputs(items []model.SplitBillItemInput, participantMap map[string]*model.Participant, taxAmount int) error {
	if len(items) == 0 {
		return model.NewAppError(nil, http.StatusBadRequest, "split_bill_items must not be empty for split_bill payment type")
	}
	if taxAmount < 0 {
		return model.NewAppError(nil, http.StatusBadRequest, "tax_amount must be greater than or equal to 0")
	}

	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			return model.NewAppError(nil, http.StatusBadRequest, "split bill item name is required")
		}
		if item.Price <= 0 {
			return model.NewAppError(nil, http.StatusBadRequest, "split bill item price must be greater than 0")
		}
		if len(item.ParticipantIDs) == 0 {
			return model.NewAppError(nil, http.StatusBadRequest, "each split bill item must have at least one assigned participant")
		}

		seen := make(map[string]struct{}, len(item.ParticipantIDs))
		for _, participantID := range item.ParticipantIDs {
			if _, ok := participantMap[participantID]; !ok {
				return model.NewAppError(nil, http.StatusBadRequest, fmt.Sprintf("participant %s is not part of this event", participantID))
			}
			if _, ok := seen[participantID]; ok {
				return model.NewAppError(nil, http.StatusBadRequest, "duplicate participant assignment found in split bill item")
			}
			seen[participantID] = struct{}{}
		}
	}

	return nil
}

func computeSplitBill(items []*model.SplitBillItem, participants map[string]*model.Participant, taxAmount int) (*splitBillComputation, error) {
	subtotalByParticipant := make(map[string]int, len(participants))
	taxByParticipant := make(map[string]int, len(participants))
	totalByParticipant := make(map[string]int, len(participants))
	itemDetails := make([]model.SplitBillItemDetail, 0, len(items))

	for participantID := range participants {
		subtotalByParticipant[participantID] = 0
		taxByParticipant[participantID] = 0
		totalByParticipant[participantID] = 0
	}

	itemsSubtotal := 0
	for _, item := range items {
		if len(item.Assignments) == 0 {
			return nil, model.NewAppError(nil, http.StatusBadRequest, "each split bill item must have at least one assigned participant")
		}

		sort.Slice(item.Assignments, func(i, j int) bool {
			return item.Assignments[i].ParticipantID < item.Assignments[j].ParticipantID
		})

		perHead := item.Price / len(item.Assignments)
		remainder := item.Price % len(item.Assignments)
		details := model.SplitBillItemDetail{
			ID:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		}

		for idx, assignment := range item.Assignments {
			share := perHead
			if idx < remainder {
				share++
			}
			subtotalByParticipant[assignment.ParticipantID] += share
			itemsSubtotal += share
			details.Assignments = append(details.Assignments, model.SplitBillItemAssignmentDetail{
				ParticipantID: assignment.ParticipantID,
				DisplayName:   participantDisplayName(participants[assignment.ParticipantID]),
				Amount:        share,
			})
		}

		itemDetails = append(itemDetails, details)
	}

	eligible := make([]string, 0)
	totalSubtotal := 0
	for participantID, subtotal := range subtotalByParticipant {
		if subtotal > 0 {
			eligible = append(eligible, participantID)
			totalSubtotal += subtotal
		}
	}
	sort.Strings(eligible)

	if taxAmount > 0 && totalSubtotal == 0 {
		return nil, model.NewAppError(nil, http.StatusBadRequest, "tax_amount cannot be distributed because no participant has assigned bill items")
	}

	type taxRemainder struct {
		ParticipantID string
		Remainder     int
	}
	remainders := make([]taxRemainder, 0, len(eligible))
	distributedTax := 0
	for _, participantID := range eligible {
		raw := taxAmount * subtotalByParticipant[participantID]
		share := raw / totalSubtotal
		rem := raw % totalSubtotal
		taxByParticipant[participantID] = share
		distributedTax += share
		remainders = append(remainders, taxRemainder{ParticipantID: participantID, Remainder: rem})
	}

	sort.Slice(remainders, func(i, j int) bool {
		if remainders[i].Remainder == remainders[j].Remainder {
			return remainders[i].ParticipantID < remainders[j].ParticipantID
		}
		return remainders[i].Remainder > remainders[j].Remainder
	})
	leftover := taxAmount - distributedTax
	for i := 0; i < leftover; i++ {
		taxByParticipant[remainders[i].ParticipantID]++
	}

	for participantID := range participants {
		totalByParticipant[participantID] = subtotalByParticipant[participantID] + taxByParticipant[participantID]
	}

	return &splitBillComputation{
		SubtotalByParticipant: subtotalByParticipant,
		TaxByParticipant:      taxByParticipant,
		TotalByParticipant:    totalByParticipant,
		ItemDetails:           itemDetails,
		ItemsSubtotal:         itemsSubtotal,
		GrandTotal:            itemsSubtotal + taxAmount,
	}, nil
}

func participantDisplayName(participant *model.Participant) string {
	if participant == nil {
		return ""
	}
	if participant.GuestName != "" {
		return participant.GuestName
	}
	return participant.User.Name
}
