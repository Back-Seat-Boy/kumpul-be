package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type paymentUsecase struct {
	paymentRepo     model.PaymentRepository
	participantRepo model.ParticipantRepository
}

func NewPaymentUsecase(paymentRepo model.PaymentRepository, participantRepo model.ParticipantRepository) model.PaymentUsecase {
	return &paymentUsecase{
		paymentRepo:     paymentRepo,
		participantRepo: participantRepo,
	}
}

func (u *paymentUsecase) GetByEventID(ctx context.Context, eventID string) (*model.Payment, error) {
	return u.paymentRepo.FindByEventID(ctx, eventID)
}

func (u *paymentUsecase) Create(ctx context.Context, eventID string, req *model.CreatePaymentRequest) (*model.Payment, error) {
	logger := log.WithFields(log.Fields{
		"context":     utils.DumpIncomingContext(ctx),
		"eventID":     eventID,
		"totalCost":   req.TotalCost,
		"paymentInfo": req.PaymentInfo,
	})

	participantCount, err := u.participantRepo.CountByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	splitAmount := req.TotalCost / int(participantCount)

	payment := &model.Payment{
		ID:          uuid.New().String(),
		EventID:     eventID,
		TotalCost:   req.TotalCost,
		SplitAmount: splitAmount,
		PaymentInfo: req.PaymentInfo,
		CreatedAt:   time.Now(),
	}

	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		logger.Error(err)
		return nil, err
	}

	return payment, nil
}
