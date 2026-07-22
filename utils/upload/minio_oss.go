package upload

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mo7amed4522/tayara_api/global"
	"go.uber.org/zap"
)

var MinioClient *Minio // 优It improves performance, but does not support dynamic configuration.

type Minio struct {
	Client *minio.Client
	bucket string
}

func GetMinio(endpoint, accessKeyID, secretAccessKey, bucketName string, useSSL bool) (*Minio, error) {
	if MinioClient != nil {
		return MinioClient, nil
	}
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL, // Set to true if using https
	})
	if err != nil {
		return nil, err
	}
	// Try creating a bucket
	err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(context.Background(), bucketName)
		if errBucketExists == nil && exists {
			// log.Printf("We already own %s\n", bucketName)
		} else {
			return nil, err
		}
	}
	MinioClient = &Minio{Client: minioClient, bucket: bucketName}
	return MinioClient, nil
}

func (m *Minio) UploadFile(file *multipart.FileHeader) (string, string, error) {
	return m.UploadFileWithPath(file, "")
}

func (m *Minio) UploadFileWithPath(file *multipart.FileHeader, customPath string) (string, string, error) {
	client, err := GetMinio(global.GVA_CONFIG.Minio.Endpoint, global.GVA_CONFIG.Minio.AccessKeyId, global.GVA_CONFIG.Minio.AccessKeySecret, global.GVA_CONFIG.Minio.BucketName, global.GVA_CONFIG.Minio.UseSSL)
	if err != nil {
		return "", "", errors.New("function GetMinio() failed, err:" + err.Error())
	}

	filename := fmt.Sprintf("%d%s", time.Now().Unix(), file.Filename)

	// Determine the key to use
	var key string
	if customPath != "" {
		key = customPath + "/" + filename
	} else {
		key = filename
	}

	f, openError := file.Open()
	if openError != nil {
		global.GVA_LOG.Error("function file.Open() failed", zap.Any("err", openError.Error()))
		return "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close() // Create file defer Close

	_, err = client.Client.PutObject(context.Background(), global.GVA_CONFIG.Minio.BucketName, key, f, file.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		global.GVA_LOG.Error("function PutObject failed", zap.Any("err", err.Error()))
		return "", "", errors.New("function PutObject failed, err:" + err.Error())
	}
	return fmt.Sprintf("%s/%s/%s", global.GVA_CONFIG.Minio.Endpoint, global.GVA_CONFIG.Minio.BucketName, key), filename, nil
}

func (m *Minio) DeleteFile(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Delete the object from MinIO
	err := m.Client.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{})
	return err
}
