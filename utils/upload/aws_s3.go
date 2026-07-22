package upload

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mo7amed4522/tayara_api/global"
	"go.uber.org/zap"
)

type AwsS3 struct{}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@object: *AwsS3
//@function: UploadFile
//@description: Upload file to Aws S3 using aws-sdk-go. See https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html#s3-examples-bucket-ops-upload-file-to-bucket
//@param: file *multipart.FileHeader
//@return: string, string, error

func (a *AwsS3) UploadFile(file *multipart.FileHeader) (string, string, error) {
	return a.UploadFileWithPath(file, "")
}

func (a *AwsS3) UploadFileWithPath(file *multipart.FileHeader, customPath string) (string, string, error) {
	session := newSession()
	uploader := s3manager.NewUploader(session)

	// Clean the filename to remove spaces and special characters
	cleanFilename := sanitizeFilename(file.Filename)
	fileKey := fmt.Sprintf("%d_%s", time.Now().Unix(), cleanFilename)

	// Determine the path to use
	var pathPrefix string
	if customPath != "" {
		pathPrefix = customPath
	} else {
		pathPrefix = global.GVA_CONFIG.AwsS3.PathPrefix
	}

	filename := pathPrefix + "/" + fileKey

	f, openError := file.Open()
	if openError != nil {
		global.GVA_LOG.Error("function file.Open() failed", zap.Any("err", openError.Error()))
		return "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close() // Create file defer Close

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(global.GVA_CONFIG.AwsS3.Bucket),
		Key:    aws.String(filename),
		Body:   f,
	})
	if err != nil {
		global.GVA_LOG.Error("function uploader.Upload() failed", zap.Any("err", err.Error()))
		return "", "", err
	}

	return global.GVA_CONFIG.AwsS3.BaseURL + "/" + filename, fileKey, nil
}

// sanitizeFilename cleans the filename by removing/replacing problematic characters
func sanitizeFilename(filename string) string {
	// Get file extension
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove or replace other problematic characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = reg.ReplaceAllString(name, "")

	// Ensure the name is not empty
	if name == "" {
		name = "file"
	}

	return name + ext
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@object: *AwsS3
//@function: DeleteFile
//@description: Delete file from Aws S3 using aws-sdk-go. See https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html#s3-examples-bucket-ops-delete-bucket-item
//@param: file *multipart.FileHeader
//@return: string, string, error

func (*AwsS3) DeleteFile(key string) error {
	session := newSession()
	svc := s3.New(session)
	filename := global.GVA_CONFIG.AwsS3.PathPrefix + "/" + key
	bucket := global.GVA_CONFIG.AwsS3.Bucket

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		global.GVA_LOG.Error("function svc.DeleteObject() failed", zap.Any("err", err.Error()))
		return errors.New("function svc.DeleteObject() failed, err:" + err.Error())
	}

	_ = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	return nil
}

// newSession Create S3 session
func newSession() *session.Session {
	sess, _ := session.NewSession(&aws.Config{
		Region:           aws.String(global.GVA_CONFIG.AwsS3.Region),
		Endpoint:         aws.String(global.GVA_CONFIG.AwsS3.Endpoint), //Minio sets the address here for compatibility.
		S3ForcePathStyle: aws.Bool(global.GVA_CONFIG.AwsS3.S3ForcePathStyle),
		DisableSSL:       aws.Bool(global.GVA_CONFIG.AwsS3.DisableSSL),
		Credentials: credentials.NewStaticCredentials(
			global.GVA_CONFIG.AwsS3.SecretID,
			global.GVA_CONFIG.AwsS3.SecretKey,
			"",
		),
	})
	return sess
}
