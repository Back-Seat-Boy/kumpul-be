package usecase

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/config"
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type uploadUsecase struct {
	cld *cloudinary.Cloudinary
}

func NewUploadUsecase(cld *cloudinary.Cloudinary) model.UploadUsecase {
	return &uploadUsecase{cld: cld}
}

func (u *uploadUsecase) UploadImage(ctx context.Context, req *model.UploadImageRequest) (*model.UploadImageResponse, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
	})

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Upload to Cloudinary
	uploadResult, err := u.cld.Upload.Upload(ctx, data, uploader.UploadParams{
		Folder: config.CloudinaryUploadFolder(),
	})
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	return &model.UploadImageResponse{
		URL:      uploadResult.SecureURL,
		PublicID: uploadResult.PublicID,
	}, nil
}
