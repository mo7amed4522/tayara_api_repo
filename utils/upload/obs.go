package upload

import (
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/mo7amed4522/tayara_api/global"

	"go.uber.org/zap"
)

var HuaWeiObs = new(Obs)

type Obs struct{}

func NewHuaWeiObsClient() (client *obs.ObsClient, err error) {
	return obs.New(global.GVA_CONFIG.HuaWeiObs.AccessKey, global.GVA_CONFIG.HuaWeiObs.SecretKey, global.GVA_CONFIG.HuaWeiObs.Endpoint)
}

func (o *Obs) UploadFile(file *multipart.FileHeader) (string, string, error) {
	return o.UploadFileWithPath(file, "")
}

func (o *Obs) UploadFileWithPath(file *multipart.FileHeader, customPath string) (string, string, error) {
	client, err := NewHuaWeiObsClient()
	if err != nil {
		return "", "", errors.New("Failed to retrieve Huawei object storage object: " + err.Error())
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
	defer f.Close()

	input := &obs.PutObjectInput{}
	input.Bucket = global.GVA_CONFIG.HuaWeiObs.Bucket
	input.Key = key
	input.Body = f

	_, err = client.PutObject(input)
	if err != nil {
		global.GVA_LOG.Error("function PutObject() failed", zap.Any("err", err.Error()))
		return "", "", errors.New("function PutObject() failed, err:" + err.Error())
	}
	return fmt.Sprintf("https://%s.%s/%s", global.GVA_CONFIG.HuaWeiObs.Bucket, global.GVA_CONFIG.HuaWeiObs.Endpoint, key), filename, nil
}

func (o *Obs) DeleteFile(key string) error {
	client, err := NewHuaWeiObsClient()
	if err != nil {
		return errors.New("Failed to retrieve Huawei object storage object: " + err.Error())
	}
	input := &obs.DeleteObjectInput{
		Bucket: global.GVA_CONFIG.HuaWeiObs.Bucket,
		Key:    key,
	}
	var output *obs.DeleteObjectOutput
	output, err = client.DeleteObject(input)
	if err != nil {
		return fmt.Errorf("Deleting object (%s) failed.!, output: %v, err: %v", key, output, err)
	}
	return nil
}
