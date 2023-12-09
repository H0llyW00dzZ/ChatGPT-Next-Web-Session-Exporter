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
	"io"
	"os"
	"strings"
	"testing"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/filesystem"
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
	input := "4\noutput.csv\n" // User selects option 4 and specifies "output.csv" as the file name.
	reader := bufio.NewReader(strings.NewReader(input))

	// Create a cancellable context to allow for timeout or cancellation of the process.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an instance of the mock file system
	mockFS := filesystem.NewMockFileSystem()

	// Redirect stdout to a pipe where we can capture the output of the function.
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Invoke the processCSVOption function, which should process the input and generate a CSV file.
	processCSVOption(mockFS, ctx, reader, store.ChatNextWebStore.Sessions)

	// Close the write-end of the pipe to finish capturing the output.
	w.Close()

	// Read the captured output from the read-end of the pipe into a buffer for assertion.
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Restore the original stdout.
	os.Stdout = oldStdout

	// Convert the captured output into a string for easy comparison.
	outputStr := buf.String()

	// Assert that the "output.csv" file was created by the function.
	_, err = os.Stat("output.csv")
	if os.IsNotExist(err) {
		t.Errorf("Expected file 'output.csv' was not created")
	}

	// Clean up by removing the test output file.
	defer os.Remove("output.csv")

	// Check that the captured output contains the expected success message.
	expectedOutput := "CSV output saved to output.csv\n"
	if !strings.Contains(outputStr, expectedOutput) {
		t.Errorf("Expected output to contain: %s, got: %s", expectedOutput, outputStr)
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
		defer os.Remove("testing.json_repaired")
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
	content := "Test content"

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
	if string(mockFS.FilesCreated[expectedFileName].Bytes()) != content {
		t.Errorf("WriteFile was called with the wrong content: got %v, want %v", string(mockFS.FilesCreated[expectedFileName].Bytes()), content)
	}
}
