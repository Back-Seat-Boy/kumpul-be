package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatusVoting      EventStatus = "voting"
	EventStatusConfirmed   EventStatus = "confirmed"
	EventStatusOpen        EventStatus = "open"
	EventStatusPaymentOpen EventStatus = "payment_open"
	EventStatusCompleted   EventStatus = "completed"
	EventStatusCancelled   EventStatus = "cancelled"
)

type EventVisibility string

const (
	EventVisibilityPublic     EventVisibility = "public"
	EventVisibilityInviteOnly EventVisibility = "invite_only"
)

type Event struct {
	ID             string          `json:"id" gorm:"primaryKey;type:uuid"`
	CreatedBy      string          `json:"created_by" gorm:"type:uuid;not null"`
	Title          string          `json:"title" gorm:"not null"`
	Description    string          `json:"description"`
	Status         EventStatus     `json:"status" gorm:"not null;default:'voting'"`
	Visibility     EventVisibility `json:"visibility" gorm:"type:varchar(20);not null;default:'invite_only'"`
	ChosenOptionID *string         `json:"chosen_option_id,omitempty" gorm:"type:uuid"`
	PlayerCap      *int            `json:"player_cap"`
	VotingDeadline *time.Time      `json:"voting_deadline"`
	ShareToken     string          `json:"share_token" gorm:"uniqueIndex;not null"`
	CreatedAt      time.Time       `json:"created_at"`
	Creator        User            `json:"creator" gorm:"foreignKey:CreatedBy"`
	ChosenOption   *EventOption    `json:"chosen_option,omitempty" gorm:"foreignKey:ChosenOptionID"`
}

type CreateEventRequest struct {
	Title                     string                      `json:"title" validate:"required"`
	Description               string                      `json:"description"`
	Visibility                EventVisibility             `json:"visibility" validate:"omitempty,oneof=public invite_only"`
	PlayerCap                 *int                        `json:"player_cap"`
	SkipVoting                bool                        `json:"skip_voting"`
	VotingDeadline            *time.Time                  `json:"voting_deadline"`
	CreateEventOptionRequests []*CreateEventOptionRequest `json:"options" validate:"required,dive"`
}

type UpdateEventStatusRequest struct {
	Status EventStatus `json:"status" validate:"required"`
}

type UpdateEventChosenOptionRequest struct {
	OptionID string `json:"option_id" validate:"required"`
}

type UpdateEventScheduleRequest struct {
	VenueID   string    `json:"venue_id" validate:"required"`
	Date      time.Time `json:"date" validate:"required"`
	StartTime string    `json:"start_time" validate:"required"`
	EndTime   string    `json:"end_time" validate:"required"`
	Note      string    `json:"note" validate:"required"`
}

// EventSummary is used for dashboard list with status-specific info
type EventSummary struct {
	Event
	// For voting status
	TotalVotes int64 `json:"total_votes,omitempty"`
	// For open/confirmed status
	ParticipantCount int64  `json:"participant_count,omitempty"`
	VenueName        string `json:"venue_name,omitempty"`
	EventDate        string `json:"event_date,omitempty"`
	EventTime        string `json:"event_time,omitempty"`
	// For payment status
	PendingCount   int64 `json:"pending_count,omitempty"`
	ClaimedCount   int64 `json:"claimed_count,omitempty"`
	ConfirmedCount int64 `json:"confirmed_count,omitempty"`
}

// ChosenOptionDetails holds venue and time details for chosen option
type ChosenOptionDetails struct {
	Date      string
	StartTime string
	EndTime   string
	VenueName string
}

// PaymentRecordCounts holds counts by status for payment records
type PaymentRecordCounts struct {
	Pending   int64
	Claimed   int64
	Confirmed int64
}

// ListEventsFilter holds filter parameters for listing events
type ListEventsFilter struct {
	Search     string          // Search in title
	Status     EventStatus     // Filter by status
	Visibility EventVisibility // Filter by visibility
	EventDate  string          // Filter by event date (YYYY-MM-DD)
}

// PaginationMode defines the pagination type
type PaginationMode string

const (
	PaginationModePage   PaginationMode = "page"
	PaginationModeCursor PaginationMode = "cursor"
)

// ListEventsRequest holds pagination and filter parameters
type ListEventsRequest struct {
	Mode            PaginationMode
	Page            int    // For page-based pagination
	Limit           int    // Page size or cursor limit
	Cursor          string // For cursor-based pagination (last event ID)
	RequesterUserID string
	Filter          ListEventsFilter
}

// ListEventsResponse is the paginated response
type ListEventsResponse struct {
	Events     []*EventSummary `json:"events"`
	Total      int64           `json:"total,omitempty"`       // Total count for page mode
	NextCursor string          `json:"next_cursor,omitempty"` // Next cursor for cursor mode
	HasMore    bool            `json:"has_more,omitempty"`    // For cursor mode
}

type EventRepository interface {
	FindByID(ctx context.Context, id string) (*Event, error)
	FindByShareToken(ctx context.Context, token string) (*Event, error)
	FindByCreatedBy(ctx context.Context, createdBy string) ([]*Event, error)
	FindByParticipantUserID(ctx context.Context, userID string) ([]*Event, error)
	FindVisibleCreatedByUser(ctx context.Context, createdBy string, requesterUserID string) ([]*Event, error)
	FindVisibleParticipatedByUser(ctx context.Context, userID string, requesterUserID string) ([]*Event, error)
	List(ctx context.Context) ([]*Event, error)
	ListPaginated(ctx context.Context, req *ListEventsRequest) ([]*Event, int64, error)
	BulkFetchVoteCounts(ctx context.Context, eventIDs []string) (map[string]int64, error)
	BulkFetchParticipantCounts(ctx context.Context, eventIDs []string) (map[string]int64, error)
	BulkFetchChosenOptionDetails(ctx context.Context, chosenOptionIDs []string) (map[string]*ChosenOptionDetails, error)
	BulkFetchPaymentRecordCounts(ctx context.Context, eventIDs []string) (map[string]*PaymentRecordCounts, error)
	Create(ctx context.Context, event *Event) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, event *Event) error
	Update(ctx context.Context, event *Event) error
	UpdateStatus(ctx context.Context, id string, status EventStatus) error
	UpdateStatusWithTx(ctx context.Context, tx *gorm.DB, id string, status EventStatus) error
	UpdateChosenOption(ctx context.Context, id string, optionID string) error
	UpdateChosenOptionWithTx(ctx context.Context, tx *gorm.DB, id string, optionID string) error
	Delete(ctx context.Context, id string) error
}

type EventUsecase interface {
	GetByID(ctx context.Context, id string) (*Event, error)
	GetByShareToken(ctx context.Context, token string) (*Event, error)
	List(ctx context.Context) ([]*Event, error)
	ListForDashboard(ctx context.Context, req *ListEventsRequest) (*ListEventsResponse, error)
	ListPublic(ctx context.Context, req *ListEventsRequest) (*ListEventsResponse, error)
	ListCreatedByUser(ctx context.Context, userID string, requesterUserID string) ([]*EventSummary, error)
	ListParticipatedByUser(ctx context.Context, userID string, requesterUserID string) ([]*EventSummary, error)
	Create(ctx context.Context, userID string, req *CreateEventRequest) (*Event, error)
	UpdateStatus(ctx context.Context, id string, status EventStatus) error
	UpdateChosenOption(ctx context.Context, id string, optionID string) error
	UpdateSchedule(ctx context.Context, id string, userID string, req *UpdateEventScheduleRequest) error
	ListScheduleChangeLogs(ctx context.Context, eventID string, userID string) ([]*EventScheduleChangeLog, error)
	// CheckAndCompleteEvent checks if all payments are confirmed and marks event as completed
	CheckAndCompleteEvent(ctx context.Context, eventID string) error
	Delete(ctx context.Context, id string) error
}
