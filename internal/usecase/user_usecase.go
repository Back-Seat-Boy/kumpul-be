package usecase

import (
	"context"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type userUsecase struct {
	userRepo model.UserRepository
}

func NewUserUsecase(userRepo model.UserRepository) model.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) GetByID(ctx context.Context, id string) (*model.User, error) {
	return u.userRepo.FindByID(ctx, id)
}

func (u *userUsecase) Create(ctx context.Context, user *model.User) error {
	return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) Update(ctx context.Context, ID string, req *model.UpdateUserInput) (*model.User, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"ID":      ID,
		"req":     utils.Dump(req),
	})
	user, err := u.userRepo.FindByID(ctx, ID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	user.Name = req.Name
	user.WhatsappNumber = req.WhatsappNumber

	if err := u.userRepo.Update(ctx, user); err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) Delete(ctx context.Context, id string) error {
	return u.userRepo.Delete(ctx, id)
}

func (u *userUsecase) List(ctx context.Context) ([]*model.User, error) {
	return u.userRepo.List(ctx)
}
