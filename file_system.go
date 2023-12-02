// @file_system.go:
// Is implementation of the FileSystem interface
// and a concrete implementation of the RealFileSystem type.
package main

import (
	"os"
)

// FileSystem is an interface that defines the methods for interacting with the file system.
type FileSystem interface {
	Create(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
}

// RealFileSystem is a concrete implementation of the FileSystem interface.
type RealFileSystem struct{}

// Create creates a new file with the given name.
// It returns a pointer to the created file and any error encountered.
func (fs RealFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// Stat returns the FileInfo structure describing the named file.
// It returns an error if the file does not exist.
func (fs RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
