package internal

import (
	"fmt"
	"os"
)

// Deleter удаляет файлы
type Deleter struct {
}

// NewDeleter конструктор
func NewDeleter() *Deleter {
	return &Deleter{}
}

// DeleteFile удаляет файл
func (d *Deleter) DeleteFile(fileToDelete string) error {
	err := os.Remove(fileToDelete)
	fmt.Printf("finish delete %s file", fileToDelete)
	return err
}
