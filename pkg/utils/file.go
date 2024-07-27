package utils

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func ReadRequestFile(file *multipart.FileHeader) (*bytes.Reader, error) {
	ogFile, err := file.Open()
	if err != nil {
		return nil, errors.New("something went wrong when opening the file")
	}

	fileBytes, err := io.ReadAll(ogFile)
	if err != nil {
		return nil, errors.New("something went wrong when opening the file")
	}

	fileReader := bytes.NewReader(fileBytes)

	return fileReader, nil
}

func CopyAndDeleteFolder(source string, destination string) error {
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destination, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, sourceFile)
		return err
	})

	if err != nil {
		return err
	}

	return nil
}
