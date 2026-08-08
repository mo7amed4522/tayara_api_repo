package upload

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/utils"
	"go.uber.org/zap"
)

var mu sync.Mutex

// maxUploadSize caps a single uploaded file (10 MB) to prevent disk-exhaustion DoS.
const maxUploadSize int64 = 10 << 20

// allowedUploadExts is the extension allowlist. Uploads outside this set are rejected
// so the storage path can never host active content (.html, .svg, .php, .sh, ...) that
// would otherwise be served back to users as stored XSS / code execution.
var allowedUploadExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".doc": true, ".docx": true,
}

func isAllowedUploadExt(ext string) bool { return allowedUploadExts[ext] }

type Local struct{}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@object: *Local
//@function: UploadFile
//@description: Upload File
//@param: file *multipart.FileHeader
//@return: string, string, error

func (*Local) UploadFile(file *multipart.FileHeader) (string, string, error) {
	return new(Local).UploadFileWithPath(file, "")
}

func (*Local) UploadFileWithPath(file *multipart.FileHeader, customPath string) (string, string, error) {
	// Read file extensions
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// Extension allowlist: reject anything not in allowedUploadExts so the storage
	// path cannot host active content that is later served back to clients.
	if !isAllowedUploadExt(ext) {
		return "", "", errors.New("file type not allowed")
	}
	// Size cap: reject empty or oversized uploads to prevent disk-exhaustion DoS.
	if file.Size <= 0 || file.Size > maxUploadSize {
		return "", "", errors.New("file size exceeds allowed limit")
	}
	// Path-traversal guard: customPath is concatenated into the on-disk path, so it
	// must be a single safe path segment (no separators, no "..").
	if customPath != "" {
		clean := filepath.Clean(customPath)
		if clean == "." || strings.HasPrefix(clean, "..") || strings.ContainsAny(customPath, "/\\") {
			return "", "", errors.New("invalid upload path")
		}
	}

	// Read filenames and encrypt them
	name := strings.TrimSuffix(file.Filename, ext)
	name = utils.MD5V([]byte(name))
	// Concatenate new filename
	filename := name + "_" + time.Now().Format("20060102150405") + ext

	// Determine the storage path
	var storePath, urlPath string
	if customPath != "" {
		storePath = global.GVA_CONFIG.Local.StorePath + "/" + customPath
		urlPath = global.GVA_CONFIG.Local.Path + "/" + customPath
	} else {
		storePath = global.GVA_CONFIG.Local.StorePath
		urlPath = global.GVA_CONFIG.Local.Path
	}

	// Try creating this path
	mkdirErr := os.MkdirAll(storePath, os.ModePerm)
	if mkdirErr != nil {
		global.GVA_LOG.Error("function os.MkdirAll() failed", zap.Any("err", mkdirErr.Error()))
		return "", "", errors.New("function os.MkdirAll() failed, err:" + mkdirErr.Error())
	}
	// Concatenate path and file name
	p := storePath + "/" + filename
	filepath := urlPath + "/" + filename

	f, openError := file.Open() // Reading files
	if openError != nil {
		global.GVA_LOG.Error("function file.Open() failed", zap.Any("err", openError.Error()))
		return "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close() // Create file defer Close

	out, createErr := os.Create(p)
	if createErr != nil {
		global.GVA_LOG.Error("function os.Create() failed", zap.Any("err", createErr.Error()))

		return "", "", errors.New("function os.Create() failed, err:" + createErr.Error())
	}
	defer out.Close() // Create file defer Close

	_, copyErr := io.Copy(out, f) // Transfer (copy) files
	if copyErr != nil {
		global.GVA_LOG.Error("function io.Copy() failed", zap.Any("err", copyErr.Error()))
		return "", "", errors.New("function io.Copy() failed, err:" + copyErr.Error())
	}
	return localPublicURL(filepath), filename, nil
}

func localPublicURL(path string) string {
	baseURL := strings.TrimRight(global.GVA_CONFIG.Local.BaseURL, "/")
	if envBase := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/"); envBase != "" {
		baseURL = envBase
	}
	if baseURL == "" {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return baseURL + path
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@object: *Local
//@function: DeleteFile
//@description: Delete file
//@param: key string
//@return: error

func (*Local) DeleteFile(key string) error {
	// Check if the key is empty.
	if key == "" {
		return errors.New("key cannot be empty")
	}

	// Verify that the key does not contain illegal characters or attempt to access files outside the storage path.
	if strings.Contains(key, "..") || strings.ContainsAny(key, `\/:*?"<>|`) {
		return errors.New("invalid key")
	}

	p := filepath.Join(global.GVA_CONFIG.Local.StorePath, key)

	// Check if the file exists
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return errors.New("The file does not exist.")
	}

	// Use file locks to prevent concurrent deletion.
	mu.Lock()
	defer mu.Unlock()

	err := os.Remove(p)
	if err != nil {
		return errors.New("File deletion failed.: " + err.Error())
	}

	return nil
}
