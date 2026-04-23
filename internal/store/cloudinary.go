package store

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryStore struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryStore(cloudName, apiKey, apiSecret string) (*CloudinaryStore, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}
	return &CloudinaryStore{client: cld}, nil
}

func (s *CloudinaryStore) UploadFile(ctx context.Context, localPath string, remoteName string) (string, error) {
	uploadResult, err := s.client.Upload.Upload(ctx, localPath, uploader.UploadParams{
		PublicID: remoteName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to cloudinary: %w", err)
	}

	_ = os.Remove(localPath)

	return uploadResult.SecureURL, nil
}