package model

import (
	"context"
	"time"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

type Participant struct {
	ID        string      `json:"id" gorm:"primaryKey;type:uuid"`
	EventID   string      `json:"event_id" gorm:"type:uuid;not null"`
	UserID    null.String `json:"user_id" gorm:"type:uuid"`
	GuestName string      `json:"guest_name" gorm:"type:varchar(255)"`
	IsGuest   bool        `json:"is_guest" gorm:"-"`
	JoinedAt  time.Time   `json:"joined_at"`
	AddedBy   null.String `json:"added_by" gorm:"type:uuid"`
	User      User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ParticipantSortOrder string

const (
	ParticipantSortOrderAsc  ParticipantSortOrder = "asc"
	ParticipantSortOrderDesc ParticipantSortOrder = "desc"
)

func (p *Participant) SetDerivedFields() {
	if p == nil {
		return
	}
	p.IsGuest = !p.UserID.Valid || p.UserID.String == ""
}

type ParticipantRepository interface {
	FindByEventID(ctx context.Context, eventID string) ([]*Participant, error)
	ListPaginatedByEvent(ctx context.Context, req *ListParticipantsRequest) ([]*Participant, int64, error)
	FindByEventIDAndID(ctx context.Context, eventID, id string) (*Participant, error)
	FindByEventIDAndUserID(ctx context.Context, eventID, userID string) (*Participant, error)
	FindByEventIDAndGuestName(ctx context.Context, eventID, guestName string) (*Participant, error)
	FindByID(ctx context.Context, id string) (*Participant, error)
	Create(ctx context.Context, participant *Participant) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, participant *Participant) error
	Delete(ctx context.Context, id string) error
	CountByEventID(ctx context.Context, eventID string) (int64, error)
}

type RemoveParticipantImpact struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
	PaidAmount    int    `json:"paid_amount"`
	NewSplit      int    `json:"new_split"`
	Difference    int    `json:"difference"`
	Action        string `json:"action"`
	ActionAmount  int    `json:"action_amount"`
}

type RemovedParticipantPayment struct {
	ParticipantID        string         `json:"participant_id"`
	DisplayName          string         `json:"display_name"`
	Status               string         `json:"status"`
	Amount               int            `json:"amount"`
	PaidAmount           int            `json:"paid_amount"`
	RefundAmount         int            `json:"refund_amount"`
	RefundID             string         `json:"refund_id,omitempty"`
	RefundStatus         string         `json:"refund_status,omitempty"`
	PendingClaimedAmount int            `json:"pending_claimed_amount"`
	Claims               []PaymentClaim `json:"claims"`
}

type RemoveParticipantResult struct {
	RemovedParticipantID string                     `json:"removed_participant_id"`
	OldSplitAmount       int                        `json:"old_split_amount"`
	NewSplitAmount       int                        `json:"new_split_amount"`
	NumPaidParticipants  int64                      `json:"num_paid_participants"`
	NumRemaining         int64                      `json:"num_remaining_participants"`
	TotalCollected       int                        `json:"total_collected"`
	TotalShouldCollect   int                        `json:"total_should_collect"`
	Difference           int                        `json:"difference"`
	RemovedPayment       *RemovedParticipantPayment `json:"removed_payment,omitempty"`
	Impacts              []RemoveParticipantImpact  `json:"impacts"`
}

type JoinAsGuestRequest struct {
	GuestName string `json:"guest_name" validate:"required"`
}

type ListParticipantsFilter struct {
	Search string
}

type ListParticipantsRequest struct {
	Mode      PaginationMode
	Page      int
	Limit     int
	Cursor    string
	EventID   string
	SortOrder ParticipantSortOrder
	Filter    ListParticipantsFilter
}

type ListParticipantsResponse struct {
	Participants []*Participant `json:"participants"`
	Total        int64          `json:"total,omitempty"`
	NextCursor   string         `json:"next_cursor,omitempty"`
	HasMore      bool           `json:"has_more,omitempty"`
}

type ParticipantUsecase interface {
	ListByEvent(ctx context.Context, req *ListParticipantsRequest) (*ListParticipantsResponse, error)
	Join(ctx context.Context, eventID string, userID string, viaShareLink bool) error
	JoinAsGuest(ctx context.Context, userID, eventID string, req *JoinAsGuestRequest, viaShareLink bool) error
	Leave(ctx context.Context, eventID string, userID string) error
	PreviewRemoveParticipant(ctx context.Context, eventID string, participantID string, requesterUserID string) (*RemoveParticipantResult, error)
	RemoveParticipant(ctx context.Context, eventID string, participantID string, requesterUserID string) (*RemoveParticipantResult, error)
	GetParticipantCount(ctx context.Context, eventID string) (int64, error)
	HandlePaymentOnJoin(ctx context.Context, eventID string, participantID string) error
	HandlePaymentOnLeave(ctx context.Context, eventID string, participantID string) error
}
