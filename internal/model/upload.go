package model

import "context"

type UploadImageRequest struct {
	ImageBase64 string `json:"image_base64" validate:"required"`
}

type UploadImageResponse struct {
	URL      string `json:"url"`
	PublicID string `json:"public_id"`
}

type UploadUsecase interface {
	UploadImage(ctx context.Context, req *UploadImageRequest) (*UploadImageResponse, error)
}
