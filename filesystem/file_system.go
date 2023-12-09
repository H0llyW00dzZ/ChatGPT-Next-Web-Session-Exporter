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
type FileSystem interface {
	Create(name string) (*os.File, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Stat(name string) (os.FileInfo, error)
}

// RealFileSystem implements the FileSystem interface by wrapping the os package functions,
// thus providing an actual file system interaction mechanism.
type RealFileSystem struct{}

// Create creates a new file with the given name.
// It wraps the os.Create function and returns a pointer to the created file along with any error encountered.
func (fs RealFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func (fs RealFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Stat returns the FileInfo structure describing the file named by the given name.
// It wraps the os.Stat function and returns the FileInfo and any error encountered, for instance,
// if the file does not exist.
func (fs RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
