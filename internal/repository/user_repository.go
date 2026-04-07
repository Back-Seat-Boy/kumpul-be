package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/config"
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/cacher"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type userRepo struct {
	db     *gorm.DB
	keeper cacher.Keeper
}

func NewUserRepository(db *gorm.DB, keeper cacher.Keeper) model.UserRepository {
	return &userRepo{db: db, keeper: keeper}
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
	})
	cacheKey := fmt.Sprintf("user:id:%s", id)

	user, mu, err := findFromCacheByKey[*model.User](r.keeper, cacheKey)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	defer cacher.SafeUnlock(mu)

	if !config.DisableCaching() {
		if user != nil {
			return user, nil
		}

		if mu == nil {
			return nil, nil
		}
	}

	err = r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	switch err {
	case nil:
		err = r.cacheUser(user)
		if err != nil {
			logger.Error(err)
		}
		return user, nil
	case gorm.ErrRecordNotFound:
		err = r.keeper.StoreNil(cacheKey)
		if err != nil {
			logger.Error(err)
		}
		return nil, model.ErrUserNotFound
	default:
		logger.Error(err)
		return nil, err
	}
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"email":   email,
	})

	cacheKey := fmt.Sprintf("user:email:%s", email)
	user, mu, err := findFromCacheByKey[*model.User](r.keeper, cacheKey)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	defer cacher.SafeUnlock(mu)

	if !config.DisableCaching() {
		if user != nil {
			return user, nil
		}

		if mu == nil {
			return nil, nil
		}
	}

	err = r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	switch err {
	case nil:
		err = r.cacheUser(user)
		if err != nil {
			logger.Error(err)
		}
		return user, nil
	case gorm.ErrRecordNotFound:
		err = r.keeper.StoreNil(cacheKey)
		if err != nil {
			logger.Error(err)
		}
		return nil, model.ErrUserNotFound
	default:
		logger.Error(err)
		return nil, err
	}
}

func (r *userRepo) FindByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"googleID": googleID,
	})
	cacheKey := fmt.Sprintf("user:google:%s", googleID)
	user, mu, err := findFromCacheByKey[*model.User](r.keeper, cacheKey)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	defer cacher.SafeUnlock(mu)

	if !config.DisableCaching() {
		if user != nil {
			return user, nil
		}

		if mu == nil {
			return nil, nil
		}
	}

	err = r.db.WithContext(ctx).First(&user, "google_id = ?", googleID).Error
	switch err {
	case nil:
		err = r.cacheUser(user)
		if err != nil {
			logger.Error(err)
		}
		return user, nil
	case gorm.ErrRecordNotFound:
		err = r.keeper.StoreNil(cacheKey)
		if err != nil {
			logger.Error(err)
		}
		return nil, model.ErrUserNotFound
	default:
		logger.Error(err)
		return nil, err
	}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":  utils.DumpIncomingContext(ctx),
			"user": utils.Dump(user),
		}).Error(err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.invalidateUserCache(user.ID, user.Email, user.GoogleID)
	return nil
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	updatesMap := map[string]interface{}{
		"name": user.Name,
	}

	if user.WhatsappNumber != "" {
		updatesMap["whatsapp_number"] = user.WhatsappNumber
	}

	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", user.ID).
		Updates(updatesMap).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":  utils.DumpIncomingContext(ctx),
			"user": utils.Dump(user),
		}).Error(err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.invalidateUserCache(user.ID, user.Email, user.GoogleID)
	return nil
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	user, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.invalidateUserCache(user.ID, user.Email, user.GoogleID)
	return nil
}

func (r *userRepo) List(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *userRepo) cacheUser(user *model.User) error {
	err := r.keeper.StoreWithoutBlocking(cacher.NewItem(fmt.Sprintf("user:id:%s", user.ID), utils.ToByte(user)))
	if err != nil {
		return err
	}
	err = r.keeper.StoreWithoutBlocking(cacher.NewItem(fmt.Sprintf("user:email:%s", user.Email), utils.ToByte(user)))
	if err != nil {
		return err
	}
	err = r.keeper.StoreWithoutBlocking(cacher.NewItem(fmt.Sprintf("user:google:%s", user.GoogleID), utils.ToByte(user)))
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepo) invalidateUserCache(id, email, googleID string) {
	keys := []string{
		fmt.Sprintf("user:id:%s", id),
		fmt.Sprintf("user:email:%s", email),
		fmt.Sprintf("user:google:%s", googleID),
	}
	r.keeper.DeleteByKeys(keys)
}
