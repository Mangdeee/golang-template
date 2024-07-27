package media

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
)

type MediaService struct {
	cfg *config.Config
}

func NewMediaService(cfg *config.Config) *MediaService {
	return &MediaService{cfg: cfg}
}

func (ms *MediaService) UploadMedia(req request.MediaUploadRequest, ctx context.Context) (string, error) {
	extension := "." + req.Filename[strings.LastIndex(req.Filename, ".")+1:]
	imageDir := utils.GetExtensionType(extension)
	imageName := uuid.Must(uuid.NewRandom()).String() + extension
	bucket := ms.cfg.S3.Bucket
	fullImagePath := imageDir + "/" + imageName

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(ms.cfg.AWS.Region)}))
	client := s3manager.NewUploader(sess)

	fileReader, err := utils.ReadRequestFile(req.File)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute*1)
	defer cancel()

	res, err := client.UploadWithContext(ctx, &s3manager.UploadInput{
		Body:   fileReader,
		Bucket: aws.String(bucket),
		Key:    aws.String(fullImagePath),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		sentry.CaptureException(err)
		return "", fmt.Errorf("upload : %w", err)
	}

	return res.Location, nil
}
