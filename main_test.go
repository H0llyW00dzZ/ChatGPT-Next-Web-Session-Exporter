// @main_test.go:
// Package main provides tests for the main package of the ChatGPT-Next-Web-Session-Exporter application.
// The tests cover loading sessions from a JSON file, processing CSV output, prompting for user input,
// and handling file operations. This ensures that the application's functionality works as expected.
package main

import (
	// Importing necessary Go standard library packages and the exporter package from the application.
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/filesystem"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/interactivity"
)

// loadTestSessions is a helper function that loads test session data from a JSON file.
// It takes a file path as an argument and returns a ChatNextWebStore instance populated with the session data,
// or an error if the file cannot be read or the data cannot be decoded.
func loadTestSessions(jsonPath string) (exporter.ChatNextWebStore, error) {
	// Open the JSON file at the provided file path.
	file, err := os.Open(jsonPath)
	if err != nil {
		return exporter.ChatNextWebStore{}, err
	}
	defer file.Close() // Ensure the file is closed when the function returns.

	var store exporter.ChatNextWebStore
	// Decode the JSON data and store it in the ChatNextWebStore struct.
	err = json.NewDecoder(file).Decode(&store)
	if err != nil {
		return exporter.ChatNextWebStore{}, err
	}

	return store, nil
}

// TestProcessCSVOption verifies the functionality of processCSVOption function.
// It simulates user input, captures the output, and checks if the CSV file is created successfully.
func TestProcessCSVOption(t *testing.T) {
	// Attempt to load session data from a JSON file to prepare for testing.
	store, err := loadTestSessions("testing.json")
	if err != nil {
		t.Fatalf("Failed to load sessions from JSON: %v", err)
	}

	// Simulate user input by creating a reader that will return the input as if typed by a user.
	// Assuming that the option for OutputFormatSeparateCSVFiles is 4 as per the constants defined in session.go.
	input := fmt.Sprintf("%d\noutput_sessions.csv\noutput_messages.csv\n", 3)
	reader := bufio.NewReader(strings.NewReader(input))

	// Create a cancellable context to allow for timeout or cancellation of the process.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an instance of the mock file system
	mockFS := filesystem.NewMockFileSystem()

	// Capture the output of the function by redirecting stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Invoke the processCSVOption function, which should process the input and generate CSV files.
	processCSVOption(mockFS, ctx, reader, store.ChatNextWebStore.Sessions)

	// Close the write-end of the pipe to finish capturing the output.
	w.Close()
	os.Stdout = oldStdout // Restore the original stdout.

	// Read the captured output from the read-end of the pipe into a buffer for assertion.
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Convert the captured output into a string for easy comparison.
	outputStr := buf.String()

	// Check that the captured output contains the expected success messages.
	expectedOutputSession := "Sessions data saved to output_sessions.csv\n"
	expectedOutputMessage := "Messages data saved to output_messages.csv\n"
	if !strings.Contains(outputStr, expectedOutputSession) {
		t.Errorf("Expected output to contain: %s, got: %s", expectedOutputSession, outputStr)
	}
	// Clean up by removing the test output files.
	defer os.Remove("output_sessions.csv")
	defer os.Remove("output_messages.csv")

	if !strings.Contains(outputStr, expectedOutputMessage) {
		t.Errorf("Expected output to contain: %s, got: %s", expectedOutputMessage, outputStr)
	}
}

// TestPromptForInput verifies that promptForInput function correctly captures and returns user input.
func TestPromptForInput(t *testing.T) {
	// Simulate user input by providing a string followed by a newline to mimic pressing Enter.
	input := "test input\n"
	reader := bufio.NewReader(strings.NewReader(input))

	// Create a cancellable context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Invoke promptForInput and capture the result.
	result, err := promptForInput(ctx, reader, "Enter input: ")
	if err != nil {
		t.Fatalf("promptForInput() returned an error: %v", err)
	}

	// Check that the captured input matches the expected simulated input.
	if result != "test input" {
		t.Errorf("promptForInput() = %q, want %q", result, "test input")
	}
}

// TestPromptForInputCancellation checks if promptForInput function respects context cancellation.
func TestPromptForInputCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	reader := bufio.NewReader(strings.NewReader("test input\n"))
	go func() {
		cancel() // Cancel the context immediately
	}()

	_, err := promptForInput(ctx, reader, "Enter input: ")
	// testing for windows now
	if err != context.Canceled && err != nil && err != io.EOF {
		t.Fatalf("Expected context.Canceled or io.EOF error, got: %v", err)
	}
}

// TestLoadTestSessionsInvalidPath verifies that loadTestSessions returns an error for non-existent files.
func TestLoadTestSessionsInvalidPath(t *testing.T) {
	// Attempt to load sessions from a non-existent file and expect an error.
	_, err := loadTestSessions("invalidpath.json")
	if err == nil {
		t.Fatalf("Expected an error for an invalid file path, got nil")
	}
}

// TestLoadIncorrectJson checks loadTestSessions' behavior when provided with incorrect JSON format.
func TestLoadIncorrectJson(t *testing.T) {
	// Attempt to load sessions from a file with incorrect JSON and expect an error.
	_, err := loadTestSessions("invalidpath.json")
	if err == nil {
		t.Fatalf("Expected an error for an invalid JSON format, got nil")
	}
}

// TestRepairJSONDataFromFile verifies the repairJSONData function with both valid and invalid file paths.
func TestRepairJSONDataFromFile(t *testing.T) {
	// Define the path to your testing.json file containing broken JSON for the test.
	brokenJSONPath := "testing.json"
	realFS := &filesystem.RealFileSystem{}

	// Test successful repair of the JSON data.
	t.Run("SuccessfulRepair", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Attempt to repair the JSON data and expect a valid file path to the repaired JSON.
		repairedPath, err := repairJSONData(realFS, ctx, brokenJSONPath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if repairedPath == "" {
			t.Errorf("Expected a file path for the repaired JSON, got an empty string")
		}

		// Read the repaired file and ensure it no longer contains errors.
		repairedContent, err := os.ReadFile(repairedPath)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(repairedContent), ",]") {
			t.Errorf("Repaired JSON still contains errors: %s", repairedContent)
		}

		// Clean up by removing the repaired file after the test.
		defer os.Remove("repaired_testing.json")
	})

	// Test repair function with a non-existent file path.
	t.Run("InvalidFilePath", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Attempt to repair JSON data from a non-existent file and expect an error.
		_, err := repairJSONData(realFS, ctx, "nonexistent.json")
		if err == nil {
			t.Errorf("Expected an error for a non-existent file path, got nil")
		}
	})
}

// TestWriteContentToFile verifies that writeContentToFile function writes the expected content to a file.
func TestWriteContentToFile(t *testing.T) {
	// Create a cancellable context to simulate context cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Simulate user input for the file name.
	var userInput bytes.Buffer
	userInput.Write([]byte("testing\n")) // Simulate user typing "testing" and pressing Enter.

	// Use the buffer as the reader for input.
	reader := bufio.NewReader(&userInput)

	// Define the content to be written to the file.
	content := `{"message": "Hello Machine ? This is a Gopher unit test."}`

	// Create a mock file system.
	mockFS := filesystem.NewMockFileSystem()

	// Invoke the function to write content to a file with "dataset" as the file type.
	writeContentToFile(mockFS, ctx, reader, content, "dataset")

	// Verify that the WriteFile method was called on the mock file system.
	if !mockFS.WriteFileCalled {
		t.Errorf("WriteFile was not called")
	}

	// Verify that the WriteFile method was called with the correct parameters.
	expectedFileName := "testing.json"
	if mockFS.WriteFilePath != expectedFileName {
		t.Errorf("WriteFile was called with the wrong file name: got %v, want %v", mockFS.WriteFilePath, expectedFileName)
	}

	// Check the content written to the mock file system.
	if string(mockFS.Files[expectedFileName]) != content {
		t.Errorf("WriteFile was called with the wrong content: got %v, want %v", string(mockFS.Files[expectedFileName]), content)
	}
}

// TestConfirmOverwrite tests the ConfirmOverwrite function from the interactivity package.
// Note: This test does not perform operations on the actual disk I/O.
func TestConfirmOverwrite(t *testing.T) {
	// Define test cases.
	tests := []struct {
		name           string
		fileExists     bool
		userInput      string
		expectedResult bool
		expectError    bool
	}{
		{"FileDoesNotExist", false, "", true, false},
		{"UserConfirmsOverwrite", true, "yes\n", true, false},
		{"UserDeniesOverwrite", true, "no\n", false, false},
		{"UserInputError", true, "", false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock file system.
			mockFS := filesystem.NewMockFileSystem()

			// Set the file existence state in the mock file system according to the test case.
			if tc.fileExists {
				// Simulate an existing file by adding it to the Files map.
				mockFS.Files["testing.json"] = []byte(`{"message": "Hello Machine ? This is a Gopher unit test."}`)
			} else {
				// Ensure the file does not exist in the Files map.
				delete(mockFS.Files, "testing.json")
			}

			// Simulate user input.
			reader := bufio.NewReader(strings.NewReader(tc.userInput))

			// Call the ConfirmOverwrite function.
			result, err := interactivity.ConfirmOverwrite(mockFS, context.Background(), reader, "testing.json")

			// Verify the result.
			if result != tc.expectedResult {
				t.Errorf("ConfirmOverwrite() = %v, want %v", result, tc.expectedResult)
			}

			// Verify the error.
			if (err != nil) != tc.expectError {
				t.Errorf("ConfirmOverwrite() error = %v, wantErr %v", err, tc.expectError)
			}
		})
	}
}
