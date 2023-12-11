// Copyright (c) 2023 H0llyW00dzZ
//
// Package filesystem provides an abstraction over the native file system operations.
//
// This allows for easy testing and mocking of file system interactions.
package filesystem

import (
	"io/fs"
	"os"
)

// FileSystem is an interface that abstracts file system operations such as creating
// files, writing to files, and retrieving file information. This allows for implementations
// that can interact with the file system or provide mock functionality for testing purposes.
// FileSystem interface now includes ReadFile method.
type FileSystem interface {
	Create(name string) (*os.File, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	ReadFile(name string) ([]byte, error) // Added ReadFile method
	Stat(name string) (os.FileInfo, error)
	FileExists(name string) (bool, error) // Added FileExists method to the interface
}

// RealFileSystem implements the FileSystem interface by wrapping the os package functions,
// thus providing an actual file system interaction mechanism.
type RealFileSystem struct{}

// Create creates a new file with the given name.
// It wraps the os.Create function and returns a pointer to the created file along with any error encountered.
func (rfs RealFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func (rfs RealFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// ReadFile reads the named file and returns the contents.
// It wraps the os.ReadFile function.
func (rfs RealFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// Stat returns the FileInfo structure describing the file named by the given name.
// It wraps the os.Stat function and returns the FileInfo and any error encountered, for instance,
// if the file does not exist.
func (rfs RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// FileExists checks if a file exists in the file system at the given path.
// It returns a boolean indicating existence, and an error for any underlying
// filesystem issues encountered.
func (rfs RealFileSystem) FileExists(name string) (bool, error) {
	_, err := rfs.Stat(name)
	if err == nil {
		return true, nil // File exists
	}
	if os.IsNotExist(err) {
		return false, nil // File does not exist
	}
	return false, err // Some other error occurred
}
