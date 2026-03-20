package usecase

import (
	"context"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type venueUsecase struct {
	venueRepo model.VenueRepository
}

func NewVenueUsecase(venueRepo model.VenueRepository) model.VenueUsecase {
	return &venueUsecase{venueRepo: venueRepo}
}

func (u *venueUsecase) GetByID(ctx context.Context, id string) (*model.Venue, error) {
	return u.venueRepo.FindByID(ctx, id)
}

func (u *venueUsecase) ListByUser(ctx context.Context, userID string) ([]*model.Venue, error) {
	return u.venueRepo.FindByCreatedBy(ctx, userID)
}

func (u *venueUsecase) Create(ctx context.Context, userID string, req *model.CreateVenueRequest) (*model.Venue, error) {
	venue := &model.Venue{
		ID:             uuid.New().String(),
		CreatedBy:      userID,
		Name:           req.Name,
		Address:        req.Address,
		WhatsappNumber: req.WhatsappNumber,
		PricePerHour:   req.PricePerHour,
		CourtCount:     req.CourtCount,
		Notes:          req.Notes,
	}

	if err := u.venueRepo.Create(ctx, venue); err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"venue": utils.Dump(venue),
		}).Error(err)
		return nil, err
	}

	return venue, nil
}

func (u *venueUsecase) Update(ctx context.Context, id string, req *model.UpdateVenueRequest) (*model.Venue, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
		"req":     utils.Dump(req),
	})

	existing, err := u.venueRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	existing.Name = req.Name
	existing.Address = req.Address
	existing.WhatsappNumber = req.WhatsappNumber
	existing.PricePerHour = req.PricePerHour
	existing.CourtCount = req.CourtCount
	existing.Notes = req.Notes

	if err := u.venueRepo.Update(ctx, existing); err != nil {
		logger.Error(err)
		return nil, err
	}

	return existing, nil
}

func (u *venueUsecase) Delete(ctx context.Context, id string) error {
	return u.venueRepo.Delete(ctx, id)
}
