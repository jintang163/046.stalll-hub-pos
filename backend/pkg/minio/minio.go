package minio

import (
	"context"
	"log"
	"stalll-hub-pos/backend/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client
var Ctx = context.Background()

func InitMinIO() {
	cfg := config.AppConfig.MinIO
	var err error
	Client, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}

	exists, err := Client.BucketExists(Ctx, cfg.Bucket)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}

	if !exists {
		err = Client.MakeBucket(Ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		log.Printf("Bucket %s created", cfg.Bucket)
	}

	log.Println("MinIO connected successfully")
}

func UploadFile(objectName string, filePath string, contentType string) error {
	cfg := config.AppConfig.MinIO
	_, err := Client.FPutObject(Ctx, cfg.Bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func GetFileURL(objectName string, expiresIn int) (string, error) {
	cfg := config.AppConfig.MinIO
	presignedURL, err := Client.PresignedGetObject(Ctx, cfg.Bucket, objectName, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
