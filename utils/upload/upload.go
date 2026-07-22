package upload

import (
	"mime/multipart"

	"github.com/mo7amed4522/tayara_api/global"
)

// OSS Object storage interface
// Author [mo7amed4522](https://github.com/mo7amed4522)
type OSS interface {
	UploadFile(file *multipart.FileHeader) (string, string, error)
	UploadFileWithPath(file *multipart.FileHeader, customPath string) (string, string, error)
	DeleteFile(key string) error
}

// NewOss OSS instantiation methods
// Author [mo7amed4522](https://github.com/mo7amed4522)
func NewOss() OSS {
	switch global.GVA_CONFIG.System.OssType {
	case "local":
		return &Local{}
	case "qiniu":
		return &Qiniu{}
	case "tencent-cos":
		return &TencentCOS{}
	case "aliyun-oss":
		return &AliyunOSS{}
	case "huawei-obs":
		return HuaWeiObs
	case "aws-s3":
		return &AwsS3{}
	case "cloudflare-r2":
		return &CloudflareR2{}
	case "minio":
		minioClient, err := GetMinio(global.GVA_CONFIG.Minio.Endpoint, global.GVA_CONFIG.Minio.AccessKeyId, global.GVA_CONFIG.Minio.AccessKeySecret, global.GVA_CONFIG.Minio.BucketName, global.GVA_CONFIG.Minio.UseSSL)
		if err != nil {
			global.GVA_LOG.Warn("You have configured to use minio, but initialization failed. Please check the availability or security configuration of minio.: " + err.Error())
			panic("Minio initialization failed.") // This is the recommended approach. If users configure miniio themselves, and an error occurs, they have to restart the service, which is also risky.
		}
		return minioClient
	default:
		return &Local{}
	}
}
