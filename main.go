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

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/repairdata"
)

// main initializes the application, setting up context for cancellation and
// starting the user interaction flow for data processing and exporting.
func main() {
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
	jsonFilePath, err := promptForInput(ctx, reader, "Enter the path to the JSON file: ")
	if err != nil {
		handleInputError(err)
		return
	}

	// Offer the user an option to repair the data before processing.
	repairData, err := promptForInput(ctx, reader, "Do you want to repair data? (yes/no): ")
	if err != nil {
		handleInputError(err)
		return
	}

	if strings.ToLower(repairData) == "yes" {
		// Attempt to repair the provided JSON data.
		// Pass the context to the repairJSONData function.
		newFilePath, err := repairJSONData(ctx, jsonFilePath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Repaired JSON data has been saved to: %s\n", newFilePath)
		os.Exit(0)
	}

	// Load and parse the JSON file into session data.
	store, err := exporter.ReadJSONFromFile(jsonFilePath)
	if err != nil {
		fmt.Printf("Error reading or parsing the JSON file: %s\n", err)
		os.Exit(1)
	}

	// Query the user for the preferred output format and process accordingly.
	outputOption, err := promptForInput(ctx, reader, "Select the output format:\n1) CSV\n2) Hugging Face Dataset\n")
	if err != nil {
		handleInputError(err)
		return
	}
	processOutputOption(ctx, reader, outputOption, store.ChatNextWebStore.Sessions)
}

// handleInputError checks the type of error and handles it accordingly.
func handleInputError(err error) {
	if err == context.Canceled || err == io.EOF {
		// Handle a context cancellation or EOF, if applicable
		fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
		os.Exit(0)
	} else {
		fmt.Printf("\n[GopherHelper] Error reading input: %s\n", err)
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
	inputChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		input, err := reader.ReadString('\n')
		if err != nil {
			errorChan <- err
		} else {
			inputChan <- input
		}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errorChan:
		return "", err
	case input := <-inputChan:
		return strings.TrimSpace(input), nil
	}
}

// processOutputOption directs the processing flow based on the user's choice of output format.
// It now respects the context for cancellation, ensuring long-running operations can be interrupted.
func processOutputOption(ctx context.Context, reader *bufio.Reader, outputOption string, sessions []exporter.Session) {
	switch outputOption {
	case "1":
		processCSVOption(ctx, reader, sessions)
	case "2":
		processDatasetOption(ctx, reader, sessions)
	default:
		fmt.Println("Invalid output option.")
	}
}

// processCSVOption prompts the user for the CSV format option and performs the corresponding actions based on the selected option.
// It takes a reader to read user input, and a slice of sessions as input.
// If the format option is 3, it prompts the user for the names of the sessions and messages CSV files to save, and calls exporter.CreateSeparateCSVFiles to create separate CSV files for sessions and messages.
// If the format option is not 3, it prompts the user for the name of the CSV file to save, and calls exporter.ConvertSessionsToCSV to convert sessions to CSV based on the selected format option.
// It prints the output file names or error messages accordingly.
func processCSVOption(ctx context.Context, reader *bufio.Reader, sessions []exporter.Session) {
	// Prompt the user for the CSV format option
	formatOptionStr, err := promptForInput(ctx, reader, "Select the message output format:\n1) Inline Formatting\n2) One Message Per Line\n3) Separate Files for Sessions and Messages\n4) JSON String in CSV\n")
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			fmt.Printf("\nError reading input: %s\n", err)
			os.Exit(1)
		}
	}

	formatOption, err := strconv.Atoi(formatOptionStr)
	if err != nil {
		// If the format option is not a valid number, print an error message and return.
		fmt.Println("\nInvalid format option.")
		return
	}

	// Execute the CSV conversion based on the selected format option.
	executeCSVConversion(ctx, formatOption, reader, sessions)
}

// processDatasetOption handles the conversion of session data to a Hugging Face Dataset format.
// It is now context-aware and will respect cancellation requests.
func processDatasetOption(ctx context.Context, reader *bufio.Reader, sessions []exporter.Session) {
	datasetOutput, err := exporter.ExtractToDataset(sessions)
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			fmt.Printf("\nError reading input: %s\n", err)
			os.Exit(1)
		}
	}
	saveToFile(ctx, reader, datasetOutput, "dataset")
}

// saveToFile prompts the user to save the provided content to a file of the specified type.
// This function now also accepts a context, allowing file operations to be cancelable.
func saveToFile(ctx context.Context, reader *bufio.Reader, content string, fileType string) {
	saveOutput, err := promptForInput(ctx, reader, fmt.Sprintf("Do you want to save the output to a file? (yes/no)\n"))
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			fmt.Printf("\nError reading input: %s\n", err)
			os.Exit(1)
		}
	}

	if strings.ToLower(saveOutput) == "yes" {
		// Collect the file name from the user and write the content to the file.
		writeContentToFile(ctx, reader, content, fileType)
	}
}

// repairJSONData attempts to repair the JSON data at the provided file path and returns the path to the repaired file.
// This function is not context-aware as it performs a single, typically quick operation.
func repairJSONData(ctx context.Context, jsonFilePath string) (string, error) {
	// Check if the context is already done before starting the operation.
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Continue if the context is not cancelled.
	}

	oldJSONBytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return "", err
	}

	// Simulate a context-aware operation (since os.ReadFile is not context-aware).
	newJSONBytes, err := repairdata.RepairSessionData(oldJSONBytes)
	if err != nil {
		return "", err
	}

	newFilePath := strings.TrimSuffix(jsonFilePath, ".json") + "_repaired.json"
	err = os.WriteFile(newFilePath, newJSONBytes, 0644)
	if err != nil {
		return "", err
	}

	return newFilePath, nil
}

// executeCSVConversion handles the CSV conversion process based on the user-selected format option.
// It is now context-aware, allowing for cancellation during the CSV conversion process.
func executeCSVConversion(ctx context.Context, formatOption int, reader *bufio.Reader, sessions []exporter.Session) {
	var csvFileName string
	var err error

	if formatOption != 3 {
		csvFileName, err = promptForInput(ctx, reader, "Enter the name of the CSV file to save: ")
		if err != nil {
			if err == context.Canceled || err == io.EOF {
				// If the error is context.Canceled or io.EOF, exit gracefully.
				fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
				os.Exit(0)
			} else {
				// For other types of errors, print the error message and exit with status code 1.
				fmt.Printf("\nError reading input: %s\n", err)
				os.Exit(1)
			}
		}
	}

	switch formatOption {
	case 3:
		// If the user chooses to create separate files, prompt for file names and execute accordingly.
		createSeparateCSVFiles(ctx, reader, sessions)
	default:
		// Otherwise, convert the sessions to a single CSV file.
		convertToSingleCSV(ctx, sessions, formatOption, csvFileName)
	}
}

// createSeparateCSVFiles prompts the user for file names and creates separate CSV files for sessions and messages.
// This function is context-aware and supports cancellation during the prompt for input.
func createSeparateCSVFiles(ctx context.Context, reader *bufio.Reader, sessions []exporter.Session) {
	sessionsFileName, err := promptForInput(ctx, reader, "Enter the name of the sessions CSV file to save: ")
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			fmt.Printf("\nError reading input: %s\n", err)
			os.Exit(1)
		}
	}

	messagesFileName, err := promptForInput(ctx, reader, "Enter the name of the messages CSV file to save: ")
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			fmt.Printf("\nError reading input: %s\n", err)
			os.Exit(1)
		}
	}

	err = exporter.CreateSeparateCSVFiles(sessions, sessionsFileName, messagesFileName)
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			fmt.Printf("\nError reading input: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Sessions data saved to %s\n", sessionsFileName)
	fmt.Printf("Messages data saved to %s\n", messagesFileName)
}

// convertToSingleCSV converts the session data to a single CSV file using the specified format option.
// It now checks for context cancellation and halts the operation if a cancellation is requested.
func convertToSingleCSV(ctx context.Context, sessions []exporter.Session, formatOption int, csvFileName string) {
	err := exporter.ConvertSessionsToCSV(ctx, sessions, formatOption, csvFileName)
	if err != nil {
		if err == context.Canceled {
			fmt.Println("Operation was canceled by the user.")
		} else {
			fmt.Printf("Failed to convert sessions to CSV: %s\n", err)
		}
		return // Handle the error as appropriate for your application
	}
	fmt.Printf("CSV output saved to %s\n", csvFileName)
}

// writeContentToFile collects a file name from the user and writes the provided content to the specified file.
// It now includes context support to handle potential cancellation during file writing.
func writeContentToFile(ctx context.Context, reader *bufio.Reader, content string, fileType string) {
	fileName, err := promptForInput(ctx, reader, fmt.Sprintf("Enter the name of the %s file to save: ", fileType))
	if err != nil {
		if err == context.Canceled || err == io.EOF {
			// If the error is context.Canceled or io.EOF, exit gracefully.
			fmt.Println("\n[GopherHelper] Exiting gracefully...\nReason: Operation canceled or end of input. Exiting program.")
			os.Exit(0)
		} else {
			// For other types of errors, print the error message and exit with status code 1.
			fmt.Printf("\nError reading input: %s\n", err)
			os.Exit(1)
		}
	}

	if fileType == "dataset" {
		fileName += ".json"
	}

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Failed to create the %s file: %s\n", fileType, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Failed to write to the %s file: %s\n", fileType, err)
		return
	}

	fmt.Printf("%s output saved to %s\n", strings.ToTitle(fileType), fileName)
}
