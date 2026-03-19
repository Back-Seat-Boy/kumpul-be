package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

type authUsecase struct {
	userRepo    model.UserRepository
	sessionUC   model.SessionUsecase
	oauthConfig *oauth2.Config
}

func NewAuthUsecase(cfg model.AuthConfig, userRepo model.UserRepository, sessionUC model.SessionUsecase) model.AuthUsecase {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &authUsecase{
		userRepo:    userRepo,
		sessionUC:   sessionUC,
		oauthConfig: oauthCfg,
	}
}

func (u *authUsecase) GetGoogleLoginURL(ctx context.Context) string {
	return u.oauthConfig.AuthCodeURL("")
}

func (u *authUsecase) HandleGoogleCallback(ctx context.Context, code string) (string, *model.User, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"code":    code,
	})
	googleUser, err := u.exchangeCodeForUser(ctx, code)
	if err != nil {
		logger.Error(err)
		return "", nil, err
	}

	user, err := u.findOrCreateUser(ctx, googleUser)
	if err != nil {
		logger.Error(err)
		return "", nil, err
	}

	session, err := u.sessionUC.Create(ctx, user)
	if err != nil {
		logger.Error(err)
		return "", nil, err
	}

	return session.ID, user, nil
}

func (u *authUsecase) Logout(ctx context.Context, sessionID string) error {
	return u.sessionUC.Delete(ctx, sessionID)
}

func (u *authUsecase) ValidateSession(ctx context.Context, sessionID string) (*model.Session, *model.User, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"sessionID": sessionID,
	})
	session, err := u.sessionUC.GetByID(ctx, sessionID)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	user, err := u.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	return session, user, nil
}

func (u *authUsecase) exchangeCodeForUser(ctx context.Context, code string) (*model.GoogleUserInfo, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"code":    code,
	})

	token, err := u.oauthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("%w: %v", model.ErrInvalidCode, err)
	}

	client := u.oauthConfig.Client(ctx, token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Errorf("failed to get user info: status %d", resp.StatusCode))
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to read user info: %w", err)
	}

	var googleUser model.GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &googleUser, nil
}

func (u *authUsecase) findOrCreateUser(ctx context.Context, googleUser *model.GoogleUserInfo) (*model.User, error) {
	logger := log.WithFields(log.Fields{
		"context":    utils.DumpIncomingContext(ctx),
		"googleUser": utils.Dump(googleUser),
	})

	user, err := u.userRepo.FindByGoogleID(ctx, googleUser.ID)
	if err == nil {
		return user, nil
	}
	if err != model.ErrUserNotFound {
		logger.Error(err)
		return nil, err
	}

	user = &model.User{
		ID:            uuid.New().String(),
		GoogleID:      googleUser.ID,
		Name:          googleUser.Name,
		Email:         googleUser.Email,
		EmailVerified: googleUser.VerifiedEmail,
		AvatarURL:     googleUser.Picture,
		Provider:      "google",
		CreatedAt:     time.Now(),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}
