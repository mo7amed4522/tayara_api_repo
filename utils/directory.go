package utils

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mo7amed4522/tayara_api/global"
	"go.uber.org/zap"
)

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: PathExists
//@description: Does the file directory exist?
//@param: path string
//@return: bool, error

func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, errors.New("file with the same name exists")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: CreateDir
//@description: Batch create folders
//@param: dirs ...string
//@return: err error

func CreateDir(dirs ...string) (err error) {
	for _, v := range dirs {
		exist, err := PathExists(v)
		if err != nil {
			return err
		}
		if !exist {
			global.GVA_LOG.Debug("create directory" + v)
			if err := os.MkdirAll(v, os.ModePerm); err != nil {
				global.GVA_LOG.Error("create directory"+v, zap.Any(" error:", err))
				return err
			}
		}
	}
	return err
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: FileMove
//@description: File movement for external access
//@param: `src` (string), `dst` (string) (`src`: source location, absolute or relative path; `dst`: destination location, absolute or relative path, must be a folder)
//@return: err error

func FileMove(src string, dst string) (err error) {
	if dst == "" {
		return nil
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}
	revoke := false
	dir := filepath.Dir(dst)
Redirect:
	_, err = os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
		if !revoke {
			revoke = true
			goto Redirect
		}
	}
	return os.Rename(src, dst)
}

func DeLFile(filePath string) error {
	return os.RemoveAll(filePath)
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: TrimSpace
//@description: Remove struct spaces
//@param: target interface (target: The target structure must be passed in as a pointer.)
//@return: null

func TrimSpace(target interface{}) {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return
	}
	t = t.Elem()
	v := reflect.ValueOf(target).Elem()
	for i := 0; i < t.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(strings.TrimSpace(v.Field(i).String()))
		}
	}
}

// FileExist Check if a file exists
func FileExist(path string) bool {
	fi, err := os.Lstat(path)
	if err == nil {
		return !fi.IsDir()
	}
	return !os.IsNotExist(err)
}
