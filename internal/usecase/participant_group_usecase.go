package usecase

import (
	"context"
	"net/http"
	"strings"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/guregu/null"
)

type participantGroupUsecase struct {
	groupRepo       model.ParticipantGroupRepository
	participantRepo model.ParticipantRepository
	eventRepo       model.EventRepository
}

func NewParticipantGroupUsecase(groupRepo model.ParticipantGroupRepository, participantRepo model.ParticipantRepository, eventRepo model.EventRepository) model.ParticipantGroupUsecase {
	return &participantGroupUsecase{
		groupRepo:       groupRepo,
		participantRepo: participantRepo,
		eventRepo:       eventRepo,
	}
}

func (u *participantGroupUsecase) List(ctx context.Context, eventID string) ([]*model.ParticipantGroup, error) {
	return u.groupRepo.ListByEventID(ctx, eventID)
}

func (u *participantGroupUsecase) Create(ctx context.Context, eventID string, requesterID string, req *model.CreateParticipantGroupRequest) (*model.ParticipantGroup, error) {
	if _, err := u.validateOrganizerAccess(ctx, eventID, requesterID); err != nil {
		return nil, err
	}

	name, err := normalizeParticipantGroupName(req.Name)
	if err != nil {
		return nil, err
	}

	groups, err := u.groupRepo.ListByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	group := &model.ParticipantGroup{
		ID:        uuid.New().String(),
		EventID:   eventID,
		Name:      name,
		SortOrder: len(groups),
	}

	if err := u.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (u *participantGroupUsecase) Update(ctx context.Context, eventID string, groupID string, requesterID string, req *model.UpdateParticipantGroupRequest) (*model.ParticipantGroup, error) {
	if _, err := u.validateOrganizerAccess(ctx, eventID, requesterID); err != nil {
		return nil, err
	}

	group, err := u.findGroupByEvent(ctx, eventID, groupID)
	if err != nil {
		return nil, err
	}

	name, err := normalizeParticipantGroupName(req.Name)
	if err != nil {
		return nil, err
	}
	group.Name = name

	if err := u.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (u *participantGroupUsecase) Delete(ctx context.Context, eventID string, groupID string, requesterID string) error {
	if _, err := u.validateOrganizerAccess(ctx, eventID, requesterID); err != nil {
		return err
	}

	group, err := u.findGroupByEvent(ctx, eventID, groupID)
	if err != nil {
		return err
	}

	return u.groupRepo.Delete(ctx, group.ID)
}

func (u *participantGroupUsecase) AssignParticipant(ctx context.Context, eventID string, participantID string, requesterID string, req *model.AssignParticipantGroupRequest) error {
	if _, err := u.validateOrganizerAccess(ctx, eventID, requesterID); err != nil {
		return err
	}

	participant, err := u.participantRepo.FindByEventIDAndID(ctx, eventID, participantID)
	if err != nil {
		return err
	}

	var groupID null.String
	if req.GroupID != nil {
		requestedGroupID := strings.TrimSpace(*req.GroupID)
		if requestedGroupID != "" {
			if _, err := u.findGroupByEvent(ctx, eventID, requestedGroupID); err != nil {
				return err
			}
			groupID = null.StringFrom(requestedGroupID)
		}
	}

	return u.participantRepo.UpdateGroupID(ctx, participant.ID, groupID)
}

func (u *participantGroupUsecase) UpdateSettings(ctx context.Context, eventID string, requesterID string, allow bool) error {
	if _, err := u.validateOrganizerAccess(ctx, eventID, requesterID); err != nil {
		return err
	}

	return u.eventRepo.UpdateAllowGroupSelfSelect(ctx, eventID, allow)
}

func (u *participantGroupUsecase) ChangeOwnGroup(ctx context.Context, eventID string, requesterID string, req *model.AssignParticipantGroupRequest) error {
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event.Status == model.EventStatusCompleted {
		return model.ErrEventAlreadyCompleted
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}
	if !event.AllowGroupSelfSelect {
		return model.ErrForbidden
	}

	participant, err := u.participantRepo.FindByEventIDAndUserID(ctx, eventID, requesterID)
	if err != nil {
		return err
	}

	var requestedGroupID string
	if req != nil && req.GroupID != nil {
		requestedGroupID = strings.TrimSpace(*req.GroupID)
	}
	if requestedGroupID == "" {
		return model.NewAppError(nil, http.StatusBadRequest, "group_id is required")
	}

	if _, err := u.findGroupByEvent(ctx, eventID, requestedGroupID); err != nil {
		return err
	}

	return u.participantRepo.UpdateGroupID(ctx, participant.ID, null.StringFrom(requestedGroupID))
}

func (u *participantGroupUsecase) validateOrganizerAccess(ctx context.Context, eventID string, requesterID string) (*model.Event, error) {
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return nil, err
	}
	if event.CreatedBy != requesterID {
		return nil, model.ErrForbidden
	}
	return event, nil
}

func (u *participantGroupUsecase) findGroupByEvent(ctx context.Context, eventID string, groupID string) (*model.ParticipantGroup, error) {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.EventID != eventID {
		return nil, model.ErrParticipantGroupNotFound
	}
	return group, nil
}

func normalizeParticipantGroupName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", model.NewAppError(nil, http.StatusBadRequest, "group name is required")
	}
	return name, nil
}
