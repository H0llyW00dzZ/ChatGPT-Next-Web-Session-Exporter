// @main.go:
// Package main is the entry point for the CLI tool that interacts with the exporter
// and repairdata packages to process and export chat session data.
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/bannercli"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/filesystem"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/interactivity"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/repairdata"
)

const (
	// Output format options
	OutputFormatCSV         = exporter.FormatOptionInline // Assuming this is the inline format
	OutputFormatDataset     = 2                           // Keep the dataset format option as is
	OutputFormatInline      = exporter.FormatOptionInline
	OutputFormatPerLine     = exporter.FormatOptionPerLine
	OutputFormatSeparateCSV = exporter.OutputFormatSeparateCSVFiles // Assuming this is the separate CSV files format
	OutputFormatJSONInCSV   = exporter.FormatOptionJSON             // Assuming this is the JSON format

	// File type
	FileTypeDataset = "dataset"

	// Prompt messages
	PromptEnterJSONFilePath        = "Enter the path to the JSON file: "
	PromptRepairData               = "Do you want to repair data? (yes/no): "
	PromptSelectOutputFormat       = "Select the output format:\n1) CSV\n2) Hugging Face Dataset\n"
	PromptSelectCSVOutputFormat    = "Select the message output format:\n1) Inline Formatting\n2) One Message Per Line\n3) JSON String in CSV\n4) Separate Files for Sessions and Messages\n"
	PromptEnterCSVFileName         = "Enter the name of the CSV file to save: "
	PromptEnterSessionsCSVFileName = "Enter the name of the sessions CSV file to save: "
	PromptEnterMessagesCSVFileName = "Enter the name of the messages CSV file to save: "
	PromptSaveOutputToFile         = "Do you want to save the output to a file? (yes/no)\n"
	PromptEnterFileName            = "Enter the name of the %s file to save: "
)

// main initializes the application, setting up context for cancellation and
// starting the user interaction flow for data processing and exporting.
func main() {
	bannercli.PrintTypingBanner("ChatGPT Session Exporter", 100*time.Millisecond)
	// Prepare a cancellable context for handling graceful shutdown.
	// This context will be passed down to functions that support cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure signal handling to gracefully terminate the application.
	// This listens for system signals like SIGINT (Ctrl+C) and terminates the application.
	setupSignalHandling(cancel)

	// Initialize a buffered reader for user input.
	reader := bufio.NewReader(os.Stdin)

	// Collect the JSON file path from the user.
	jsonFilePath, err := promptForInput(ctx, reader, PromptEnterJSONFilePath)
	if err != nil {
		handleInputError(err)
		return
	}

	// Offer the user an option to repair the data before processing.
	repairData, err := promptForInput(ctx, reader, PromptRepairData)
	if err != nil {
		handleInputError(err)
		return
	}

	if strings.ToLower(repairData) == "yes" {
		// Create an instance of your real file system implementation.
		realFS := &filesystem.RealFileSystem{}
		// Pass the real file system instance when calling repairJSONData.
		newFilePath, err := repairJSONData(realFS, ctx, jsonFilePath)
		if err != nil {
			errorMessage := fmt.Sprintf("Error: %s\n", err)
			bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
			os.Exit(1)
		}
		successMessage := fmt.Sprintf("Repaired JSON data has been saved to: %s\n", newFilePath)
		bannercli.PrintTypingBanner(successMessage, 100*time.Millisecond)
		os.Exit(0)
	}

	// Load and parse the JSON file into session data.
	store, err := exporter.ReadJSONFromFile(jsonFilePath)
	if err != nil {
		errorMessage := fmt.Sprintf("Error reading or parsing the JSON file: %s\n", err)
		bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
		os.Exit(1)
	}

	// Query the user for the preferred output format and process accordingly.
	outputOption, err := promptForInput(ctx, reader, PromptSelectOutputFormat)
	if err != nil {
		handleInputError(err)
		return
	}

	// Create an instance of your real file system implementation.
	realFS := &filesystem.RealFileSystem{}
	// Pass the real file system instance when calling processOutputOption.
	processOutputOption(realFS, ctx, reader, outputOption, store.ChatNextWebStore.Sessions)
}

// handleInputError checks the type of error and handles it accordingly.
func handleInputError(err error) {
	if err == context.Canceled || err == io.EOF {
		// Handle a context cancellation or EOF, if applicable
		bannercli.PrintTypingBanner("\nReason: Operation canceled or end of input. Exiting program.", 100*time.Millisecond)
		os.Exit(0)
	} else {
		// Format the error message before passing it to PrintTypingBanner
		errorMessage := fmt.Sprintf("\n[GopherHelper] Error reading input: %s\n", err)
		bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
		os.Exit(1)
	}
}

// setupSignalHandling configures the application to respond to interrupt signals for
// graceful shutdown. It utilizes the provided cancel function to terminate operations
// when an interrupt signal (SIGINT) or termination signal (SIGTERM) is received.
// The function uses a goroutine and a channel to listen for these signals, ensuring
// that the signal handling does not block the main execution flow of the program.
func setupSignalHandling(cancel context.CancelFunc) {
	// Prepare a channel to listen for system interrupt signals.
	signals := make(chan os.Signal, 1)
	// Register the channel to receive notification of SIGINT and SIGTERM signals.
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	// Start a new goroutine that will block waiting for a signal.
	go func() {
		<-signals // Wait for a signal
		fmt.Println("\n[GopherHelper] Exiting gracefully...")
		cancel() // Cancel the context
	}()
}

// promptForInput displays a prompt to the user and returns the trimmed input response.
// It supports context cancellation, which can interrupt the blocking read operation.
func promptForInput(ctx context.Context, reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	type result struct {
		input string
		err   error
	}
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

// processOutputOption directs the processing flow based on the user's choice of output format.
// It now respects the context for cancellation, ensuring long-running operations can be interrupted.
func processOutputOption(fs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, outputOption string, sessions []exporter.Session) {
	switch outputOption {
	case `1`:
		processCSVOption(fs, ctx, reader, sessions)
	case `2`:
		processDatasetOption(fs, ctx, reader, sessions)
	default:
		bannercli.PrintTypingBanner("\nInvalid output option.", 100*time.Millisecond)
	}
}

// processCSVOption prompts the user for the CSV format option and performs the corresponding actions based on the selected option.
// It takes a reader to read user input, and a slice of sessions as input.
// If the format option is 3, it prompts the user for the names of the sessions and messages CSV files to save, and calls exporter.CreateSeparateCSVFiles to create separate CSV files for sessions and messages.
// If the format option is not 3, it prompts the user for the name of the CSV file to save, and calls exporter.ConvertSessionsToCSV to convert sessions to CSV based on the selected format option.
// It prints the output file names or error messages accordingly.
func processCSVOption(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, sessions []exporter.Session) {
	// Prompt the user for the CSV format option
	formatOptionStr, err := promptForInput(ctx, reader, PromptSelectCSVOutputFormat)
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			bannercli.PrintTypingBanner("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.", 100*time.Millisecond)
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			errorMessage := fmt.Sprintf("\n[GopherHelper] Error reading input: %s\n", err)
			bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
			os.Exit(1)
		}
	}

	formatOption, err := strconv.Atoi(formatOptionStr)
	if err != nil {
		// If the format option is not a valid number, print an error message and return.
		bannercli.PrintTypingBanner("\nInvalid format option.", 100*time.Millisecond)
		return
	}

	// Execute the CSV conversion based on the selected format option.
	executeCSVConversion(rfs, ctx, reader, formatOption, sessions)
}

// processDatasetOption handles the conversion of session data to a Hugging Face Dataset format.
// It is now context-aware and will respect cancellation requests.
func processDatasetOption(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, sessions []exporter.Session) {
	datasetOutput, err := exporter.ExtractToDataset(sessions)
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			bannercli.PrintTypingBanner("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.", 100*time.Millisecond)
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			errorMessage := fmt.Sprintf("\n[GopherHelper] Error reading input: %s\n", err)
			bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
			os.Exit(1)
		}
	}
	saveToFile(rfs, ctx, reader, datasetOutput, "dataset")
}

// saveToFile prompts the user to save the provided content to a file of the specified type.
// This function now also accepts a context, allowing file operations to be cancelable.
func saveToFile(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, content string, fileType string) {
	// Ask user if they want to save the output to a file
	saveOutput, err := promptForInput(ctx, reader, PromptSaveOutputToFile)
	if err != nil {
		handleInputError(err)
		return
	}

	if strings.ToLower(saveOutput) == "yes" {
		// Determine the file name here (or pass it as a parameter)
		fileName, err := promptForInput(ctx, reader, fmt.Sprintf(PromptEnterFileName, fileType))
		if err != nil {
			handleInputError(err)
			return
		}

		// Ensure the fileName is not empty
		if fileName == "" {
			bannercli.PrintTypingBanner("No file name entered. Operation cancelled.", 100*time.Millisecond)
			return
		}

		// Append the appropriate file extension based on the fileType
		if fileType == FileTypeDataset {
			fileName += ".json"
		} else {
			fileName += ".csv" // Assuming default fileType is CSV
		}

		// Check if the file exists and confirm overwrite if necessary
		overwrite, err := interactivity.ConfirmOverwrite(rfs, ctx, reader, fileName)
		if err != nil {
			handleInputError(err)
			return
		}
		if !overwrite {
			bannercli.PrintTypingBanner("Operation cancelled by the user.", 100*time.Millisecond)
			return
		}

		// Now that we've confirmed, attempt to write the file
		err = rfs.WriteFile(fileName, []byte(content), 0644)
		if err != nil {
			errorMessage := fmt.Sprintf("Error writing file: %s", err)
			bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
			return
		}

		successMessage := fmt.Sprintf("%s output saved to %s", strings.ToTitle(fileType), fileName)
		bannercli.PrintTypingBanner(successMessage, 100*time.Millisecond)
	} else {
		bannercli.PrintTypingBanner("Save to file operation cancelled by the user.", 100*time.Millisecond)
	}
}

// handleInputCancellation checks the error type and handles context cancellation and EOF.
func handleInputCancellation(err error) {
	if err == context.Canceled || err == io.EOF {
		bannercli.PrintTypingBanner("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.", 100*time.Millisecond)
		os.Exit(0)
	} else {
		errorMessage := fmt.Sprintf("\nError reading input: %s\n", err)
		bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
		os.Exit(1)
	}
}

// repairJSONData attempts to repair malformed JSON data at the provided file path.
// Despite accepting a context parameter, it currently does not support cancellation.
// The function reads the broken JSON, repairs it, and writes the repaired JSON back to a new file.
func repairJSONData(rfs filesystem.FileSystem, ctx context.Context, jsonFilePath string) (string, error) {
	// Read the broken JSON data using the file system interface
	data, err := rfs.ReadFile(jsonFilePath)
	if err != nil {
		return "", err // Handle the error properly
	}

	// Repair the JSON data (this is where you fix the JSON string)
	repairedData, repairErr := repairdata.RepairSessionData(data)
	if repairErr != nil {
		return "", repairErr // Handle the error properly
	}

	// Define the path for the repaired file
	repairedPath := "repaired_" + jsonFilePath

	// Write the repaired JSON data using the file system interface
	err = rfs.WriteFile(repairedPath, repairedData, 0644)
	if err != nil {
		return "", err // Handle the error properly
	}

	// Return the path to the repaired file
	return repairedPath, nil
}

// executeCSVConversion handles the CSV conversion process based on the user-selected format option.
// It is now context-aware, allowing for cancellation during the CSV conversion process.
func executeCSVConversion(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, formatOption int, sessions []exporter.Session) {
	var csvFileName string
	var err error

	// Check if the format option is valid before proceeding
	if formatOption != OutputFormatInline && formatOption != OutputFormatPerLine &&
		formatOption != OutputFormatSeparateCSV && formatOption != OutputFormatJSONInCSV {
		bannercli.PrintTypingBanner("Invalid CSV format option.", 100*time.Millisecond)
		return
	}

	// If the format option is not for separate CSV files, prompt for a single CSV file name.
	if formatOption != OutputFormatSeparateCSV {
		csvFileName, err = promptForInput(ctx, reader, PromptEnterCSVFileName)
		if err != nil {
			handleInputError(err)
			return
		}
	}

	switch formatOption {
	case OutputFormatInline, OutputFormatPerLine, OutputFormatJSONInCSV:
		// Call the function to convert sessions to a single CSV file
		convertToSingleCSV(rfs, ctx, reader, sessions, formatOption, csvFileName)
	case OutputFormatSeparateCSV:
		// Call the function to create separate CSV files for sessions and messages
		createSeparateCSVFiles(rfs, ctx, reader, sessions)
	default:
		bannercli.PrintTypingBanner("Invalid format option.", 100*time.Millisecond)
	}
}

// createSeparateCSVFiles prompts the user for file names and creates separate CSV files for sessions and messages.
// This function is context-aware and supports cancellation during the prompt for input.
func createSeparateCSVFiles(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, sessions []exporter.Session) {
	sessionsFileName, err := promptForInput(ctx, reader, PromptEnterSessionsCSVFileName)
	if err != nil {
		handleInputError(err)
		return
	}

	// Confirm overwrite for sessions CSV file
	overwrite, err := interactivity.ConfirmOverwrite(rfs, ctx, reader, sessionsFileName)
	if err != nil {
		handleInputError(err)
		return
	}
	if !overwrite {
		bannercli.PrintTypingBanner("Operation cancelled by the user for sessions file.", 100*time.Millisecond)
		return
	}

	messagesFileName, err := promptForInput(ctx, reader, PromptEnterMessagesCSVFileName)
	if err != nil {
		handleInputError(err)
		return
	}

	// Confirm overwrite for messages CSV file
	overwrite, err = interactivity.ConfirmOverwrite(rfs, ctx, reader, messagesFileName)
	if err != nil {
		handleInputError(err)
		return
	}
	if !overwrite {
		bannercli.PrintTypingBanner("Operation cancelled by the user for messages file.", 100*time.Millisecond)
		return
	}

	err = exporter.CreateSeparateCSVFiles(sessions, sessionsFileName, messagesFileName)
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			bannercli.PrintTypingBanner("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.", 100*time.Millisecond)
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			errorMessage := fmt.Sprintf("\nError creating CSV files: %s\n", err)
			bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
			os.Exit(1)
		}
	}

	successMessageSessions := fmt.Sprintf("Sessions data saved to %s\n", sessionsFileName)
	bannercli.PrintTypingBanner(successMessageSessions, 100*time.Millisecond)

	successMessageMessages := fmt.Sprintf("Messages data saved to %s\n", messagesFileName)
	bannercli.PrintTypingBanner(successMessageMessages, 100*time.Millisecond)
}

// convertToSingleCSV converts the session data to a single CSV file using the specified format option.
// It now checks for context cancellation and halts the operation if a cancellation is requested.
func convertToSingleCSV(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, sessions []exporter.Session, formatOption int, csvFileName string) {
	// Confirm overwrite if the file already exists
	overwrite, err := interactivity.ConfirmOverwrite(rfs, ctx, reader, csvFileName)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to check file existence: %s\n", err)
		bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
		return // Handle the error as appropriate for your application
	}
	if !overwrite {
		bannercli.PrintTypingBanner("Operation cancelled by the user.", 100*time.Millisecond)
		return
	}

	err = exporter.ConvertSessionsToCSV(ctx, sessions, formatOption, csvFileName)
	if err != nil {
		if err == context.Canceled {
			bannercli.PrintTypingBanner("Operation was canceled by the user.", 100*time.Millisecond)
		} else {
			errorMessage := fmt.Sprintf("Failed to convert sessions to CSV: %s\n", err)
			bannercli.PrintTypingBanner(errorMessage, 100*time.Millisecond)
		}
		return // Handle the error as appropriate for your application
	}

	successMessage := fmt.Sprintf("CSV output saved to %s\n", csvFileName)
	bannercli.PrintTypingBanner(successMessage, 100*time.Millisecond)
}

// writeContentToFile collects a file name from the user and writes the provided content to the specified file.
// It now includes context support to handle potential cancellation during file writing.
// Note: Do not refactor or modify this function; doing so will disrupt the associated magic method in main_test.go.
func writeContentToFile(rfs filesystem.FileSystem, ctx context.Context, reader *bufio.Reader, content string, fileType string) error {
	fileName, err := promptForInput(ctx, reader, fmt.Sprintf(PromptEnterFileName, fileType))
	if err != nil {
		return err
	}

	if fileType == "dataset" {
		fileName += ".json"
	}

	// Use the provided FileSystem interface to write the file content directly
	err = rfs.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		return err
	}

	successMessage := fmt.Sprintf("%s output saved to %s\n", strings.ToTitle(fileType), fileName)
	bannercli.PrintTypingBanner(successMessage, 100*time.Millisecond)
	return nil // Ensure that you return nil if there were no errors
}
