// Below, the package filesystem (@file_system_mock.go) furnishes a mock implementation of the FileSystem
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
	"unsafe" // this package is used to convert MockFile as Expert in the Real World.

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
)

// Ensure MockFileSystem adheres to the FileSystem interface.
var _ FileSystem = (*MockFileSystem)(nil)

// MockFileSystem is a mock implementation of the FileSystem interface that can be used in tests.
// It uses a map to store file names and associated data, allowing for the simulation of file creation,
// reading, and writing without actual file system interaction.
type MockFileSystem struct {
	Files                 map[string][]byte // Files maps file names to file contents.
	WriteFileCalled       bool              // Track if WriteFile has been called.
	WriteFilePath         string            // Track the path provided to WriteFile.
	WriteFileData         []byte            // Track the data provided to WriteFile.
	WriteFilePerm         fs.FileMode       // Track the file permissions provided to WriteFile.
	FileExistsCalled      bool              // Track if FileExists has been called.
	FileExistsErr         error             // Track the error to return from FileExists.
	FileExistsShouldError bool              // Track if FileExists should return an error.
	ReadFileCalled        bool              // this field to track if ReadFile has been caled.
	ReadFileData          []byte            // Optionally track the data provided to ReadFile.
	ReadFileErr           error             // Optionally track the error provider to ReadFile.
}

// MockExporter is a mock implementation of the exporter.Exporter interface for testing purposes.
// It allows for the simulation of session conversion to CSV format and can be set to return errors for testing error handling.
//
// Note: this types is proof of concept after touring golang.
type MockExporter struct {
	ErrToReturn error // ErrToReturn is the error that ConvertSessionsToCSV will return when called.
}

// mockFileInfo is a dummy implementation of fs.FileInfo used for testing.
// It provides basic implementations of the fs.FileInfo interface methods.
type mockFileInfo struct {
	name string // name is the file name.
	*bytes.Buffer
}

// MockFile simulates an os.File for testing purposes, may you not using this if you are not Expert.
type MockFile struct {
	name     string
	contents *bytes.Buffer
}

// FileLike is an interface that includes the subset of *os.File methods that you need to mock.
type FileLike interface {
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
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
		Files: make(map[string][]byte),
	}
}

// Stat returns the FileInfo for the given file name if it exists in the mock file system.
// If the file does not exist, it returns an error to simulate the os.Stat behavior.
func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	if _, ok := m.Files[name]; ok {
		// Return mock file information.
		return mockFileInfo{name: name}, nil
	}
	return nil, os.ErrNotExist
}

// Create simulates the creation of a file by adding a new entry in the Files map.
func (m *MockFileSystem) Create(name string) (*os.File, error) {
	if _, exists := m.Files[name]; exists {
		return nil, os.ErrExist
	}
	m.Files[name] = []byte{}
	mockFile := &MockFile{
		name:     name,
		contents: bytes.NewBuffer([]byte{}),
	}
	// You would need to convert MockFile to *os.File using an interface or other means.
	// This is a simplified example and may not work directly without additional setup.
	var file *os.File = (*os.File)(unsafe.Pointer(mockFile)) // Unsafe conversion for example purposes.
	return file, nil
}

// ReadFile simulates reading the content of a file from the Files map.
// If the file exists, it returns the content as a byte slice; otherwise, it returns an error.
func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if content, ok := m.Files[name]; ok {
		return content, nil
	}
	return nil, fs.ErrNotExist
}

// WriteFile simulates writing data to a file in the Files map.
// It creates a new buffer with the provided data, simulating a successful write operation.
func (m *MockFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.Files[name] = data
	m.WriteFileCalled = true // Set this to true when WriteFile is called.
	m.WriteFilePath = name   // Record the path.
	m.WriteFileData = data   // Record the data.
	m.WriteFilePerm = perm   // Record the permissions.
	return nil
}

// FileExists checks if the given file name exists in the mock file system.
func (m *MockFileSystem) FileExists(name string) (bool, error) {
	m.FileExistsCalled = true // Record that FileExists was called
	if m.FileExistsShouldError {
		return false, m.FileExistsErr
	}
	_, exists := m.Files[name]
	return exists, nil
}

func (mf *MockFileSystem) Close() error {
	// Implement the Close method if needed for testing
	return nil
}

// Close simulates closing the file, it's a no-op for the mock.
func (mf *MockFile) Close() error {
	return nil // No-op for the mock.
}

// Write simulates writing bytes to the file.
func (mf *MockFile) Write(p []byte) (n int, err error) {
	return mf.contents.Write(p)
}

// Read simulates reading bytes from the file.
func (mf *MockFile) Read(p []byte) (n int, err error) {
	return mf.contents.Read(p)
}

// Seek simulates seeking in the file, it's a no-op for the mock.
func (mf *MockFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil // No-op for the mock.
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }           // Dummy value for size.
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }           // Dummy value for file mode.
func (m mockFileInfo) ModTime() time.Time { return time.Time{} } // Dummy value for modification time.
func (m mockFileInfo) IsDir() bool        { return false }       // Dummy value, always false.
func (m mockFileInfo) Sys() interface{}   { return nil }         // No system-specific information.
