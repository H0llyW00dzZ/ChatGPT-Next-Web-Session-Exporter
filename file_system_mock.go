// @file_system_mock.go:
// A mock implementation of the FileSystem interface for testing purposes.
package main

import (
	"io/fs"
	"os"
)

// MockFileSystem is a mock implementation of the FileSystem interface for testing purposes.
type MockFileSystem struct {
	FilesCreated []string
}

// Stat implements FileSystem.
// Stat returns the FileInfo for the given file name.
// It is currently unimplemented and will panic if called.
func (*MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	panic("unimplemented")
}

// Create simulates creating a file and records the action.
// It takes a file name as input and returns a dummy file.
// The file name is added to the FilesCreated slice for tracking purposes.
func (m *MockFileSystem) Create(name string) (*os.File, error) {
	m.FilesCreated = append(m.FilesCreated, name)
	// You can return a real file or a dummy file here depending on your testing needs.
	// Below we're returning a dummy file to avoid actual file system interaction.
	return os.NewFile(0, name), nil
}

// WasCalledWith checks if Create was called with the given filename.
// It iterates over the FilesCreated slice and returns true if the given filename is found.
// Otherwise, it returns false.
func (m *MockFileSystem) WasCalledWith(name string) bool {
	for _, fileCreated := range m.FilesCreated {
		if fileCreated == name {
			return true
		}
	}
	return false
}
