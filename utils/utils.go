package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetExecutableName 获取当前可执行文件的名称
func GetExecutableName() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Base(exePath), nil
}

// IsDirWritable 判断目录是否可写
func IsDirWritable(dir string) bool {
	testFile := filepath.Join(dir, ".writetest")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	return true
}

func GenPid(dir string) error {
	executeName, err := GetExecutableName()
	if err != nil {
		return err
	}
	if !IsDirWritable(dir) {
		return fmt.Errorf("directory is not writable: %s", dir)
	}
	fileName := executeName + ".pid"
	pidFilePath := filepath.Join(dir, fileName)
	pid := os.Getpid()
	file, err := os.Create(pidFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = fmt.Fprint(file, pid); err != nil {
		return err
	}
	return nil
}
