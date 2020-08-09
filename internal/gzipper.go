package internal

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GzipExtension расширение для архивного файла
const GzipExtension = ".gz"

// Gzipper сжимает файлы
type Gzipper struct {
	FileName string
}

// NewGzipper конструктор
func NewGzipper() *Gzipper {
	return &Gzipper{
		FileName: "",
	}
}

// GzipFile сжимает заданный файл
func (g *Gzipper) GzipFile(source string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}

	filename := filepath.Base(source)
	target := source + GzipExtension
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filename
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)

	g.FileName = target
	fmt.Printf("finish gzip %s\n", g.FileName)
	return err
}
