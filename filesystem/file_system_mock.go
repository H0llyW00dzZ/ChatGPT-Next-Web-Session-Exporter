// Below, the package filesystem furnishes a mock implementation of the FileSystem
// interface intended for testing. It enables the monitoring of file operations and
// the emulation of file system interactions without real disk I/O, showcasing the
// "magic" of Go programming language (🎩🪄).
//
// Additionally, it can be utilized for tracking file activities during tests or without tests (as Expert).
//
// Copyright (c) 2023 H0llyW00dzZ
package filesystem

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"time"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
)

// Ensure MockFileSystem adheres to the FileSystem interface.
var _ FileSystem = (*MockFileSystem)(nil)

// MockFileSystem is a mock implementation of the FileSystem interface that can be used in tests.
// It uses a map to store file names and associated data, allowing for the simulation of file creation,
// reading, and writing without actual file system interaction.
type MockFileSystem struct {
	FilesCreated          map[string]*bytes.Buffer // FilesCreated maps file names to buffers holding file contents.
	WriteFileCalled       bool                     // Add this field to track if WriteFile has been called.
	WriteFilePath         string                   // Track the path provided to WriteFile.
	WriteFileData         []byte                   // Optionally track the data provided to WriteFile.
	WriteFilePerm         fs.FileMode              // Optionally track the file permissions provided to WriteFile.
	Files                 map[string][]byte        // Files maps file names to file contents.
	ReadFileCalled        bool                     // this field to track if ReadFile has been caled.
	ReadFileData          []byte                   // Optionally track the data provided to ReadFile.
	ReadFileErr           error                    // Optionally track the error provider to ReadFile.
	FileExistsCalled      bool                     // Optionally track the result of FileExists.
	FileExistsErr         error                    // Optionally track the error provider to FileExists.
	FileExistsShouldError bool                     // Optionally track if FileExists should return an error.
}

// MockExporter is a mock implementation of the exporter.Exporter interface for testing purposes.
// It allows for the simulation of session conversion to CSV format and can be set to return errors for testing error handling.
//
// Note: this types is proof of concept after touring golang.
type MockExporter struct {
	ErrToReturn error // ErrToReturn is the error that ConvertSessionsToCSV will return when called.
}

// ConvertSessionsToCSV simulates the conversion of sessions to CSV format.
// It returns an error specified by ErrToReturn, allowing for error handling tests.
//
// Note: this function is proof of concept after touring golang.
func (m *MockExporter) ConvertSessionsToCSV(ctx context.Context, sessions []exporter.Session, formatOption int, csvFileName string) error {
	return m.ErrToReturn
}

// NewMockFileSystem creates a new instance of MockFileSystem with initialized internal structures.
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		FilesCreated: make(map[string]*bytes.Buffer),
		Files:        make(map[string][]byte),
	}
}

// Stat returns the FileInfo for the given file name if it exists in the mock file system.
// If the file does not exist, it returns an error to simulate the os.Stat behavior.
func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	if _, ok := m.FilesCreated[name]; ok {
		// Return mock file information. In real tests, you might want to return a more detailed mock.
		return mockFileInfo{name: name}, nil
	}
	return nil, fs.ErrNotExist
}

// Create simulates the creation of a file by creating a new buffer in the FilesCreated map.
// It returns nil to signify that the file is not meant for actual I/O operations.
func (m *MockFileSystem) Create(name string) (*os.File, error) {
	// Now Safe to create a new buffer for the file.
	m.FilesCreated[name] = new(bytes.Buffer)
	// Return nil to indicate that the file is not meant for I/O operations.
	return nil, nil
}

// ReadFile simulates reading the content of a file from the Files map.
// If the file exists, it returns the content as a byte slice; otherwise, it returns an error.
func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if content, ok := m.Files[name]; ok {
		return content, nil
	}
	return nil, fs.ErrNotExist
}

// WriteFile simulates writing data to a file in the FilesCreated map.
// It creates a new buffer with the provided data, simulating a successful write operation.
func (m *MockFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if m.FilesCreated == nil {
		m.Files = make(map[string][]byte)
	}
	m.FilesCreated[name] = bytes.NewBuffer(data)
	m.WriteFileCalled = true // Set this to true when WriteFile is called.
	m.WriteFilePath = name   // Record the path.
	m.WriteFileData = data   // Record the data.
	m.WriteFilePerm = perm   // Record the permissions.
	return nil
}

// mockFileInfo is a dummy implementation of fs.FileInfo used for testing.
// It provides basic implementations of the fs.FileInfo interface methods.
type mockFileInfo struct {
	name string // name is the file name.
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }           // Dummy value for size.
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }           // Dummy value for file mode.
func (m mockFileInfo) ModTime() time.Time { return time.Time{} } // Dummy value for modification time.
func (m mockFileInfo) IsDir() bool        { return false }       // Dummy value, always false.
func (m mockFileInfo) Sys() interface{}   { return nil }         // No system-specific information.

// FileExists checks if the given file name exists in the mock file system.
func (m *MockFileSystem) FileExists(name string) (bool, error) {
	m.FileExistsCalled = true // Record that FileExists was called
	if m.FileExistsErr != nil {
		// Simulate an error condition if an error is set
		return false, nil
	}
	_, exists := m.Files[name] // Use the same map for all file data which can easily be mocked while touring in binary.
	return exists, nil
}
