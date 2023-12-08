// @main_test.go:
// Is Package main provides tests for the main package of the ChatGPT-Next-Web-Session-Exporter application.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
)

// loadTestSessions loads the test sessions from a JSON file.
// It takes the path to the JSON file as input and returns a ChatNextWebStore and an error.
func loadTestSessions(jsonPath string) (exporter.ChatNextWebStore, error) {
	file, err := os.Open(jsonPath)
	if err != nil {
		return exporter.ChatNextWebStore{}, err
	}
	defer file.Close()

	var store exporter.ChatNextWebStore
	err = json.NewDecoder(file).Decode(&store)
	if err != nil {
		return exporter.ChatNextWebStore{}, err
	}

	return store, nil
}

// TestProcessCSVOption tests the processCSVOption function.
// It loads the session data from a JSON file, mocks user input, captures the standard output,
// calls the function being tested, and asserts the expected output.
func TestProcessCSVOption(t *testing.T) {
	// Load the session data from the JSON file
	store, err := loadTestSessions("testing.json")
	if err != nil {
		t.Fatalf("Failed to load sessions from JSON: %v", err)
	}

	// Mock user input: select inline formatting and provide a file name
	input := "4\noutput.csv\n"
	reader := bufio.NewReader(strings.NewReader(input))

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture the original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function being tested
	processCSVOption(ctx, reader, store.ChatNextWebStore.Sessions)

	// Close the write-end of the pipe so we can read from it
	w.Close()

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Restore os.Stdout
	os.Stdout = oldStdout

	outputStr := buf.String()

	// Check if the file was created
	_, err = os.Stat("output.csv")
	if os.IsNotExist(err) {
		t.Errorf("Expected file 'output.csv' was not created")
	}

	// Clean up test file
	defer os.Remove("output.csv")

	// Assert the expected output
	expectedOutput := "CSV output saved to output.csv\n"
	if !strings.Contains(outputStr, expectedOutput) {
		t.Errorf("Expected output to contain: %s, got: %s", expectedOutput, outputStr)
	}
}

// TestPromptForInput tests the promptForInput function.
// It mocks user input, calls the function being tested, and asserts the expected result.
func TestPromptForInput(t *testing.T) {
	input := "test input\n"
	reader := bufio.NewReader(strings.NewReader(input))

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call promptForInput and capture the result
	result, err := promptForInput(ctx, reader, "Enter input: ")
	if err != nil {
		t.Fatalf("promptForInput() returned an error: %v", err)
	}

	// Assert the expected result
	if result != "test input" {
		t.Errorf("promptForInput() = %q, want %q", result, "test input")
	}
}
