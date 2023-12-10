// Package interactivity provides functions to handle interactive command-line
// user interfaces. It includes utilities for prompting the user with questions
// and processing their responses. This package is specifically designed to work
// with the filesystem provided by the github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/filesystem
// package to confirm potential file overwrites and to collect user input in a
// context-aware manner, allowing for graceful cancellation of input requests.
package interactivity

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/filesystem"
)

// result is a helper struct used internally within the interactivity package
// to encapsulate the user input along with any error that might have occurred
// during the input reading process. It is used to communicate between
// goroutines in the promptForInput function.
type result struct {
	input string
	err   error
}

// ConfirmOverwrite checks if a file with the given fileName exists in the provided filesystem.
// If the file does exist, it prompts the user for confirmation to overwrite the file.
// The function reads the user's input via the provided bufio.Reader and expects a 'yes' or 'no' response.
// A context.Context is used to handle cancellation of the input request.
// It returns a boolean indicating whether the file should be overwritten and any error encountered.
func ConfirmOverwrite(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, fileName string) (bool, error) {
	exists, err := rfs.FileExists(fileName)
	if err != nil {
		// Handle the error properly, perhaps by returning it.
		return false, err
	}
	if !exists {
		// If the file doesn't exist, no need to confirm overwrite.
		return true, nil
	}

	// If the file exists, ask the user for confirmation.
	fmt.Printf("File '%s' already exists. Overwrite? (yes/no): ", fileName)

	// Call promptForInput without the extra string argument.
	overwrite, err := promptForInput(ctx, reader)
	if err != nil {
		return false, err
	}
	return strings.ToLower(overwrite) == "yes", nil
}

// promptForInput waits for a line of user input read from the provided bufio.Reader.
// It takes a context.Context to support cancellation.
// The function trims the newline character from the input and returns the resulting string.
// If the context is cancelled before the user inputs a line, the context's error is returned.
func promptForInput(ctx context.Context, reader *bufio.Reader) (string, error) {
	resultChan := make(chan result)

	go func() {
		input, err := reader.ReadString('\n')
		resultChan <- result{input: input, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultChan:
		return strings.TrimSpace(res.input), res.err
	}
}
