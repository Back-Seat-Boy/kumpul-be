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

func (u *venueUsecase) ListAll(ctx context.Context) ([]*model.Venue, error) {
	return u.venueRepo.ListAll(ctx)
}

func (u *venueUsecase) ListPaginated(ctx context.Context, req *model.ListVenuesRequest) (*model.ListVenuesResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	venues, total, err := u.venueRepo.ListPaginated(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(venues) == 0 {
		return &model.ListVenuesResponse{
			Venues:  []*model.Venue{},
			Total:   total,
			HasMore: false,
		}, nil
	}

	nextCursor := ""
	hasMore := false

	if len(venues) > 0 {
		lastVenue := venues[len(venues)-1]
		nextCursor = lastVenue.ID

		if req.Mode == model.PaginationModeCursor {
			hasMore = int64(len(venues)) == int64(req.Limit) && total > int64(len(venues))
		} else {
			offset := req.Page * req.Limit
			hasMore = int64(offset) < total
		}
	}

	return &model.ListVenuesResponse{
		Venues:     venues,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (u *venueUsecase) Create(ctx context.Context, userID string, req *model.CreateVenueRequest) (*model.Venue, error) {
	venue := &model.Venue{
		ID:             uuid.New().String(),
		CreatedBy:      userID,
		Name:           req.Name,
		Address:        req.Address,
		WhatsappNumber: req.WhatsappNumber,
		MapsURL:        req.MapsURL,
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
	existing.MapsURL = req.MapsURL
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
