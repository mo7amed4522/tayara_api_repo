package upload

import (
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mo7amed4522/tayara_api/global"
	"go.uber.org/zap"
)

type AliyunOSS struct{}

func (*AliyunOSS) UploadFile(file *multipart.FileHeader) (string, string, error) {
	return new(AliyunOSS).UploadFileWithPath(file, "")
}

func (*AliyunOSS) UploadFileWithPath(file *multipart.FileHeader, customPath string) (string, string, error) {
	bucket, err := NewBucket()
	if err != nil {
		global.GVA_LOG.Error("function AliyunOSS.NewBucket() failed", zap.Any("err", err.Error()))
		return "", "", errors.New("function AliyunOSS.NewBucket() failed, err:" + err.Error())
	}

	// Read local files.
	f, openError := file.Open()
	if openError != nil {
		global.GVA_LOG.Error("function file.Open() failed", zap.Any("err", openError.Error()))
		return "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close() // Create file defer Close
	// Upload path to Alibaba Cloud. You can change the filename format, but it's recommended to ensure uniqueness.
	yunFileTmpPath := fmt.Sprintf("%d%s", time.Now().Unix(), file.Filename)

	// Determine the path to use
	var objectKey string
	if customPath != "" {
		objectKey = customPath + "/" + yunFileTmpPath
	} else {
		objectKey = yunFileTmpPath
	}

	// Upload file stream.
	err = bucket.PutObject(objectKey, f)
	if err != nil {
		global.GVA_LOG.Error("function formUploader.Put() failed", zap.Any("err", err.Error()))
		return "", "", errors.New("function formUploader.Put() failed, err:" + err.Error())
	}

	return global.GVA_CONFIG.AliyunOSS.BucketUrl + "/" + objectKey, yunFileTmpPath, nil
}

func (*AliyunOSS) DeleteFile(key string) error {
	bucket, err := NewBucket()
	if err != nil {
		global.GVA_LOG.Error("function AliyunOSS.NewBucket() Failed", zap.Any("err", err.Error()))
		return errors.New("function AliyunOSS.NewBucket() Failed, err:" + err.Error())
	}

	// Delete a single file. `objectName` indicates that a complete path, including the file extension, must be specified when deleting an OSS file, for example, `abc/efg/123.jpg`.。
	// To delete a folder, set objectName to the corresponding folder name. If the folder is not empty, you need to delete all objects within the folder before you can delete the folder itself.。
	err = bucket.DeleteObject(key)
	if err != nil {
		global.GVA_LOG.Error("function bucketManager.Delete() failed", zap.Any("err", err.Error()))
		return errors.New("function bucketManager.Delete() failed, err:" + err.Error())
	}

	return nil
}

func NewBucket() (*oss.Bucket, error) {
	// Create an OSSClient instance.
	client, err := oss.New(global.GVA_CONFIG.AliyunOSS.Endpoint, global.GVA_CONFIG.AliyunOSS.AccessKeyId, global.GVA_CONFIG.AliyunOSS.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	// Obtain storage space.
	bucket, err := client.Bucket(global.GVA_CONFIG.AliyunOSS.BucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
