package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"container/list"
)

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateFileDir(dir string) {
	exist, err := PathExists(dir)
	if err != nil {
		return
	}

	if !exist {
		os.MkdirAll(dir, os.ModePerm)
	}
}

func GetCurrentPath() (string, error) {
	return os.Getwd()
}

func ReadFile(filenm string) ([]byte, error) {
	file, err := os.OpenFile(filenm, os.O_RDONLY, 0766)
	if err != nil {
		MyPrintf("ReadFile failed:%s,%s.\r\n", filenm, err.Error())
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func WriteFile(filenm string, inbuf []byte) (int, error) {
	file, err := os.OpenFile(filenm, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		MyPrintf("WriteFile failed:%s,%s.\r\n", filenm, err.Error())
		return 0, err
	}

	defer file.Close()

	return file.Write(inbuf)
}

func RemoveFile(filenm string) bool {
	err := os.Remove(filenm)
	if err != nil {
		FileLogs.Info("RemoveFile failed:%s-%s.\r\n", filenm, err.Error())
		return false
	}

	return true
}

func MoveFile(filenm string, newpath string) {
	err := os.Rename(filenm, newpath)
	if err != nil {
		MyPrintf("moveFile failed:%s-%s.\r\n", filenm, newpath, err.Error())
	}
}

func GetFilelist(path string) *list.List {
	filelst := list.New()
	lstlen := filelst.Len()

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		filelst.PushBack(path)
		lstlen = filelst.Len()
		if lstlen > 100 {
			return errors.New("maxnums exit")
		}

		return nil
	})

	if err != nil && lstlen <= 0 {
		return nil
	}

	if lstlen > 0 {
		//MyPrintf("GetFilelist:%d.\r\n", lstlen)
	}

	return filelst
}

func FileSize(fpath string) (int64, error) {
	f, e := os.Stat(fpath)
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}
