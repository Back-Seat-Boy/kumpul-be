package model

import (
	"context"
	"time"
)

type ParticipantGroup struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventID   string    `json:"event_id" gorm:"type:uuid;not null;index"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateParticipantGroupRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateParticipantGroupRequest struct {
	Name string `json:"name" validate:"required"`
}

type AssignParticipantGroupRequest struct {
	GroupID *string `json:"group_id"`
}

type UpdateParticipantGroupSettingsRequest struct {
	AllowGroupSelfSelect bool `json:"allow_group_self_select"`
}

type ParticipantGroupRepository interface {
	ListByEventID(ctx context.Context, eventID string) ([]*ParticipantGroup, error)
	FindByID(ctx context.Context, id string) (*ParticipantGroup, error)
	Create(ctx context.Context, group *ParticipantGroup) error
	Update(ctx context.Context, group *ParticipantGroup) error
	Delete(ctx context.Context, id string) error
}

type ParticipantGroupUsecase interface {
	List(ctx context.Context, eventID string) ([]*ParticipantGroup, error)
	Create(ctx context.Context, eventID string, requesterID string, req *CreateParticipantGroupRequest) (*ParticipantGroup, error)
	Update(ctx context.Context, eventID string, groupID string, requesterID string, req *UpdateParticipantGroupRequest) (*ParticipantGroup, error)
	Delete(ctx context.Context, eventID string, groupID string, requesterID string) error
	AssignParticipant(ctx context.Context, eventID string, participantID string, requesterID string, req *AssignParticipantGroupRequest) error
	UpdateSettings(ctx context.Context, eventID string, requesterID string, allow bool) error
	ChangeOwnGroup(ctx context.Context, eventID string, requesterID string, req *AssignParticipantGroupRequest) error
}
