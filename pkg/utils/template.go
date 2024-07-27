package utils

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

func ParseTemplateDir(dir string, file string) (*template.Template, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == "base.html" || info.Name() == "styles.html" || info.Name() == file) {
			paths = append(paths, path)
		}
		return nil
	})

	fmt.Println("Parsing templates...")

	if err != nil {
		return nil, err
	}

	return template.ParseFiles(paths...)
}
