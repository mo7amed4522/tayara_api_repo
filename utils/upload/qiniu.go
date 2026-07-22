package upload

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"

	"go.uber.org/zap"
)

type Qiniu struct{}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@object: *Qiniu
//@function: UploadFile
//@description: Upload File
//@param: file *multipart.FileHeader
//@return: string, string, error

func (*Qiniu) UploadFile(file *multipart.FileHeader) (string, string, error) {
	return new(Qiniu).UploadFileWithPath(file, "")
}

func (*Qiniu) UploadFileWithPath(file *multipart.FileHeader, customPath string) (string, string, error) {
	putPolicy := storage.PutPolicy{Scope: global.GVA_CONFIG.Qiniu.Bucket}
	mac := qbox.NewMac(global.GVA_CONFIG.Qiniu.AccessKey, global.GVA_CONFIG.Qiniu.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{UseHTTPS: global.GVA_CONFIG.Qiniu.UseHTTPS, UseCdnDomains: global.GVA_CONFIG.Qiniu.UseCdnDomains}
	formUploader := storage.NewFormUploader(&cfg)

	ret := storage.PutRet{}
	putExtra := storage.PutExtra{Params: map[string]string{"x:name": "github logo"}}

	f, openError := file.Open()
	if openError != nil {
		global.GVA_LOG.Error("function file.Open() failed", zap.Any("err", openError.Error()))
		return "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close()                                                  // Create file defer Close
	fileKey := fmt.Sprintf("%d%s", time.Now().Unix(), file.Filename) // You can change the file name format, but it's recommended to ensure uniqueness.

	// Determine the key to use
	var key string
	if customPath != "" {
		key = customPath + "/" + fileKey
	} else {
		key = fileKey
	}

	putErr := formUploader.Put(context.Background(), &ret, upToken, key, f, file.Size, &putExtra)
	if putErr != nil {
		global.GVA_LOG.Error("function formUploader.Put() failed", zap.Any("err", putErr.Error()))
		return "", "", errors.New("function formUploader.Put() failed, err:" + putErr.Error())
	}
	return global.GVA_CONFIG.Qiniu.ImgPath + "/" + ret.Key, ret.Key, nil
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@object: *Qiniu
//@function: DeleteFile
//@description: Delete file
//@param: key string
//@return: error

func (*Qiniu) DeleteFile(key string) error {
	mac := qbox.NewMac(global.GVA_CONFIG.Qiniu.AccessKey, global.GVA_CONFIG.Qiniu.SecretKey)
	cfg := qiniuConfig()
	bucketManager := storage.NewBucketManager(mac, cfg)
	if err := bucketManager.Delete(global.GVA_CONFIG.Qiniu.Bucket, key); err != nil {
		global.GVA_LOG.Error("function bucketManager.Delete() failed", zap.Any("err", err.Error()))
		return errors.New("function bucketManager.Delete() failed, err:" + err.Error())
	}
	return nil
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@object: *Qiniu
//@function: qiniuConfig
//@description: Configure the return to Qiniu Cloud according to the configuration file.
//@return: *storage.Config

func qiniuConfig() *storage.Config {
	cfg := storage.Config{
		UseHTTPS:      global.GVA_CONFIG.Qiniu.UseHTTPS,
		UseCdnDomains: global.GVA_CONFIG.Qiniu.UseCdnDomains,
	}
	switch global.GVA_CONFIG.Qiniu.Zone { // Initialize the corresponding data center according to the configuration file.
	case "ZoneHuadong":
		cfg.Zone = &storage.ZoneHuadong
	case "ZoneHuabei":
		cfg.Zone = &storage.ZoneHuabei
	case "ZoneHuanan":
		cfg.Zone = &storage.ZoneHuanan
	case "ZoneBeimei":
		cfg.Zone = &storage.ZoneBeimei
	case "ZoneXinjiapo":
		cfg.Zone = &storage.ZoneXinjiapo
	}
	return &cfg
}
