// Package filesystem provides a mock implementation of the FileSystem interface for testing.
package filesystem

import (
	"io/fs"
	"os"
)

// MockFileSystem is a mock implementation of the FileSystem interface that can be used in tests.
// It tracks the file operations performed, such as file creation, without interacting with the actual file system.
type MockFileSystem struct {
	FilesCreated []string // FilesCreated keeps track of the names of the files that have been "created".
}

// Stat is part of the FileSystem interface. It is meant to return the FileInfo for the given file name.
// In the MockFileSystem, this function is unimplemented and will panic if called, indicating the function
// should be stubbed or the test should be structured to avoid its call.
func (*MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	panic("unimplemented")
}

// Create simulates the creation of a file by adding the provided file name to the FilesCreated slice.
// It returns a dummy *os.File object that can be used in place of a real file in tests.
// This allows tests to verify that a file creation operation was initiated without actually creating a file on disk.
func (m *MockFileSystem) Create(name string) (*os.File, error) {
	m.FilesCreated = append(m.FilesCreated, name)
	// Here we return a dummy file using os.NewFile. The file descriptor 0 signifies that this is not associated
	// with an open file. The name is used for identification in tests.
	return os.NewFile(0, name), nil
}

// WasCalledWith checks whether the Create method was called with a specific file name.
// It returns true if the name is in the FilesCreated slice, indicating that a create operation was requested
// for the file; otherwise, it returns false. This is useful for assertions in tests to ensure that
// the correct file operations were performed.
func (m *MockFileSystem) WasCalledWith(name string) bool {
	for _, fileCreated := range m.FilesCreated {
		if fileCreated == name {
			return true
		}
	}
	return false
}
