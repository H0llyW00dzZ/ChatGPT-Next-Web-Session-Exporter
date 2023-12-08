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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure signal handling to gracefully terminate the application.
	setupSignalHandling(cancel)

	// Initialize a buffered reader for user input.
	reader := bufio.NewReader(os.Stdin)

	// Collect the JSON file path from the user.
	jsonFilePath := promptForInput(reader, "Enter the path to the JSON file: ")

	// Offer the user an option to repair the data before processing.
	repairData := promptForInput(reader, "Do you want to repair data? (yes/no): ")
	if strings.ToLower(repairData) == "yes" {
		// Attempt to repair the provided JSON data.
		newFilePath, err := repairJSONData(jsonFilePath)
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
	outputOption := promptForInput(reader, "Select the output format:\n1) CSV\n2) Hugging Face Dataset\n")
	processOutputOption(ctx, reader, outputOption, store.ChatNextWebStore.Sessions)
}

// setupSignalHandling configures the application to respond to interrupt signals for
// graceful shutdown, utilizing the provided cancel function to terminate operations.
func setupSignalHandling(cancel context.CancelFunc) {
	// Prepare a channel to listen for system interrupt signals.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		fmt.Println("Signal received, cancelling operations...")
		cancel()
	}()
}

// promptForInput displays a prompt to the user and returns the trimmed input response.
func promptForInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("EOF encountered. Exiting program.")
		} else {
			fmt.Printf("Error reading input: %s\n", err)
		}
		os.Exit(1)
	}
	return strings.TrimSpace(input)
}

// processOutputOption directs the processing flow based on the user's choice of output format.
func processOutputOption(ctx context.Context, reader *bufio.Reader, outputOption string, sessions []exporter.Session) {
	switch outputOption {
	case "1":
		processCSVOption(ctx, reader, sessions)
	case "2":
		processDatasetOption(reader, sessions)
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
	formatOptionStr := promptForInput(reader, "Select the message output format:\n1) Inline Formatting\n2) One Message Per Line\n3) Separate Files for Sessions and Messages\n4) JSON String in CSV\n")
	formatOption, err := strconv.Atoi(formatOptionStr)
	if err != nil {
		fmt.Println("Invalid format option.")
		return
	}

	// Execute the CSV conversion based on the selected format option.
	executeCSVConversion(ctx, formatOption, reader, sessions)
}

// processDatasetOption handles the conversion of session data to a Hugging Face Dataset format.
func processDatasetOption(reader *bufio.Reader, sessions []exporter.Session) {
	datasetOutput, err := exporter.ExtractToDataset(sessions)
	if err != nil {
		fmt.Printf("Failed to extract to dataset: %s\n", err)
		return
	}
	saveToFile(reader, datasetOutput, "dataset")
}

// saveToFile prompts the user to save the provided content to a file of the specified type.
func saveToFile(reader *bufio.Reader, content string, fileType string) {
	saveOutput := promptForInput(reader, fmt.Sprintf("Do you want to save the output to a file? (yes/no)\n"))
	if saveOutput == "yes" {
		// Collect the file name from the user and write the content to the file.
		writeContentToFile(reader, content, fileType)
	}
}

// repairJSONData attempts to repair the JSON data at the provided file path and returns the path to the repaired file.
func repairJSONData(jsonFilePath string) (string, error) {
	oldJSONBytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return "", fmt.Errorf("reading the JSON file: %w", err)
	}

	newJSONBytes, err := repairdata.RepairSessionData(oldJSONBytes)
	if err != nil {
		return "", fmt.Errorf("repairing the JSON data: %w", err)
	}

	newFilePath := strings.TrimSuffix(jsonFilePath, ".json") + "_repaired.json"
	err = os.WriteFile(newFilePath, newJSONBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("writing the new JSON data to file: %w", err)
	}

	return newFilePath, nil
}

// executeCSVConversion handles the CSV conversion process based on the user-selected format option.
func executeCSVConversion(ctx context.Context, formatOption int, reader *bufio.Reader, sessions []exporter.Session) {
	csvFileName := ""
	if formatOption != 3 {
		csvFileName = promptForInput(reader, "Enter the name of the CSV file to save: ")
	}

	switch formatOption {
	case 3:
		// If the user chooses to create separate files, prompt for file names and execute accordingly.
		createSeparateCSVFiles(reader, sessions)
	default:
		// Otherwise, convert the sessions to a single CSV file.
		convertToSingleCSV(ctx, sessions, formatOption, csvFileName)
	}
}

// createSeparateCSVFiles prompts the user for file names and creates separate CSV files for sessions and messages.
func createSeparateCSVFiles(reader *bufio.Reader, sessions []exporter.Session) {
	sessionsFileName := promptForInput(reader, "Enter the name of the sessions CSV file to save: ")
	messagesFileName := promptForInput(reader, "Enter the name of the messages CSV file to save: ")

	err := exporter.CreateSeparateCSVFiles(sessions, sessionsFileName, messagesFileName)
	if err != nil {
		fmt.Printf("Failed to create separate CSV files: %s\n", err)
		return
	}

	fmt.Printf("Sessions data saved to %s\n", sessionsFileName)
	fmt.Printf("Messages data saved to %s\n", messagesFileName)
}

// convertToSingleCSV converts the session data to a single CSV file using the specified format option.
func convertToSingleCSV(ctx context.Context, sessions []exporter.Session, formatOption int, csvFileName string) {
	err := exporter.ConvertSessionsToCSV(ctx, sessions, formatOption, csvFileName)
	if err != nil {
		if err == context.Canceled {
			fmt.Println("Operation was canceled by the user.")
		} else {
			fmt.Printf("Failed to convert sessions to CSV: %s\n", err)
		}
		return
	}
	fmt.Printf("CSV output saved to %s\n", csvFileName)
}

// writeContentToFile collects a file name from the user and writes the provided content to the specified file.
func writeContentToFile(reader *bufio.Reader, content string, fileType string) {
	fileName := promptForInput(reader, fmt.Sprintf("Enter the name of the %s file to save: ", fileType))
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
