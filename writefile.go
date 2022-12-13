package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func writeFileToDisk(fileName string, contents []byte, mode fs.FileMode) (n int, err error) {
	fileDir := filepath.Dir(fileName)
	fileBase := filepath.Base(fileName)
	fileTemp := filepath.Join(fileDir, ".tmp."+fileBase+"~")

	err = os.MkdirAll(fileDir, 0o777)
	if err != nil {
		return n, fmt.Errorf("os.MkdirAll: %q: %w", fileDir, err)
	}

	dir, err := os.OpenFile(fileDir, os.O_RDONLY, 0)
	if err != nil {
		return n, fmt.Errorf("os.OpenFile: %q, O_RDONLY: %w", fileDir, err)
	}

	needDirClose := true
	defer func() {
		if needDirClose {
			_ = dir.Close()
		}
	}()

	_ = os.Remove(fileTemp)

	file, err := os.OpenFile(fileTemp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return n, fmt.Errorf("os.OpenFile: %q, O_WRONLY|O_CREATE|O_EXCL: %w", fileTemp, err)
	}

	needFileClose := true
	needFileRemove := true
	defer func() {
		if needFileClose {
			_ = file.Close()
		}
		if needFileRemove {
			_ = os.Remove(fileTemp)
		}
	}()

	n, err = file.Write(contents)
	if err != nil {
		return n, fmt.Errorf("os.File.Write: %q: %w", fileTemp, err)
	}

	err = file.Sync()
	if err != nil {
		return n, fmt.Errorf("os.File.Sync: %q: %w", fileTemp, err)
	}

	needFileClose = false
	err = file.Close()
	if err != nil {
		return n, fmt.Errorf("os.File.Close: %q: %w", fileTemp, err)
	}

	err = os.Rename(fileTemp, fileName)
	if err != nil {
		return n, fmt.Errorf("os.Rename: %q, %q: %w", fileTemp, fileName, err)
	}

	needFileRemove = false
	err = dir.Sync()
	if err != nil {
		return n, fmt.Errorf("os.File.Sync: %q: %w", fileDir, err)
	}

	needDirClose = false
	err = dir.Close()
	if err != nil {
		return n, fmt.Errorf("os.File.Close: %q: %w", fileDir, err)
	}

	return n, nil
}
