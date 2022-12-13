package indexutil

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

func WriteFile(ctx context.Context, filePath string, contents []byte, mode fs.FileMode) {
	logger := zerolog.Ctx(ctx)

	filePath = filepath.Clean(filePath)
	dirPath := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	tempPath := filepath.Join(dirPath, ".tmp."+baseName+"~")

	modeName := fmt.Sprintf("%03o", mode)

	// Preserve "r" and "w" bits, and set "x" bit to equal "r" bit.
	dirMode := (0o666 & mode) | ((0o444 & mode) >> 2)
	dirModeName := fmt.Sprintf("%03o", dirMode)

	err := os.MkdirAll(dirPath, dirMode)
	if err != nil {
		logger.Fatal().
			Str("path", dirPath).
			Str("mode", dirModeName).
			Err(err).
			Msg("failed to create parent directory")
		panic(nil)
	}

	dir, err := os.OpenFile(dirPath, os.O_RDONLY, 0)
	if err != nil {
		logger.Fatal().
			Str("path", dirPath).
			Err(err).
			Msg("failed to open parent directory for metadata sync")
		panic(nil)
	}

	needDirClose := true
	defer func() {
		if needDirClose {
			_ = dir.Close()
		}
	}()

	_ = os.Remove(tempPath)

	file, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		logger.Fatal().
			Str("path", tempPath).
			Str("mode", modeName).
			Err(err).
			Msg("failed to create temporary file")
		panic(nil)
	}

	needFileClose := true
	needFileRemove := true
	defer func() {
		if needFileClose {
			_ = file.Close()
		}
		if needFileRemove {
			_ = os.Remove(tempPath)
		}
	}()

	_, err = file.Write(contents)
	if err != nil {
		logger.Fatal().
			Str("path", tempPath).
			Err(err).
			Msg("I/O error while writing to temporary file")
		panic(nil)
	}

	err = file.Sync()
	if err != nil {
		logger.Fatal().
			Str("path", tempPath).
			Err(err).
			Msg("I/O error while syncing file data to disk")
		panic(nil)
	}

	needFileClose = false
	err = file.Close()
	if err != nil {
		logger.Fatal().
			Str("path", tempPath).
			Err(err).
			Msg("I/O error while closing file")
		panic(nil)
	}

	err = os.Rename(tempPath, filePath)
	if err != nil {
		logger.Fatal().
			Str("old", tempPath).
			Str("new", filePath).
			Err(err).
			Msg("failed to rename file to permanent filename")
		panic(nil)
	}

	needFileRemove = false
	err = dir.Sync()
	if err != nil {
		logger.Fatal().
			Str("path", dirPath).
			Err(err).
			Msg("I/O error while syncing file metadata to disk")
		panic(nil)
	}

	needDirClose = false
	err = dir.Close()
	if err != nil {
		logger.Fatal().
			Str("path", dirPath).
			Err(err).
			Msg("I/O error while closing parent directory")
		panic(nil)
	}
}
