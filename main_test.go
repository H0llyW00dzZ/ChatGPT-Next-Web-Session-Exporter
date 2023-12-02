package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
)

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
func TestProcessCSVOption(t *testing.T) {
	// Load the session data from the JSON file
	sessions, err := loadTestSessions("testing.json")
	if err != nil {
		t.Fatalf("Failed to load sessions from JSON: %v", err)
	}

	// Mock user input: select inline formatting and provide a file name
	input := "4\noutput.csv\n"
	reader := bufio.NewReader(strings.NewReader(input))

	// Create a pipe to capture the standard output
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }() // Restore original Stdout

	outputBuffer := bytes.Buffer{}

	// Call the function being tested
	go func() {
		processCSVOption(reader, sessions.ChatNextWebStore.Sessions)
		w.Close()
	}()

	// Read from the pipe into the buffer
	_, err = outputBuffer.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	// Get the output from the buffer
	outputStr := outputBuffer.String()

	// Check if the file was created
	_, err = os.Stat("output.csv")
	if os.IsNotExist(err) {
		t.Errorf("Expected file 'output.csv' was not created")
	}

	// Clean up test file
	os.Remove("output.csv")

	// Assert the expected output
	expectedOutput := "CSV output saved to output.csv\n"
	if !strings.Contains(outputStr, expectedOutput) {
		t.Errorf("Expected output to contain: %s, but got: %s", expectedOutput, outputStr)
	}
}

// TestPromptForInput tests the promptForInput function.
func TestPromptForInput(t *testing.T) {
	input := "test input\n"
	reader := bufio.NewReader(strings.NewReader(input))

	result := promptForInput(reader, "Enter input: ")

	if result != "test input" {
		t.Errorf("promptForInput() = %q, want %q", result, "test input")
	}
}
