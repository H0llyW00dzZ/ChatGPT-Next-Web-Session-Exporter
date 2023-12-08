// @main.go:
// Package main is the entry point for the CLI tool that utilizes the exporter package
// to process chat session data and provide various output formats.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/repairdata"
)

// main is the entry point of the CLI tool.
func main() {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling to cancel the context on SIGINT or SIGTERM
	setupSignalHandling(cancel)

	reader := bufio.NewReader(os.Stdin)

	// Get the JSON file path from the user
	jsonFilePath := promptForInput(reader, "Enter the path to the JSON file: ")

	// Prompt the user to decide whether to repair data
	repairData := promptForInput(reader, "Do you want to repair data? (yes/no): ")
	if strings.ToLower(repairData) == "yes" {
		// Read the old JSON content from the file
		oldJSONBytes, err := os.ReadFile(jsonFilePath)
		if err != nil {
			fmt.Printf("Error reading the JSON file: %s\n", err)
			os.Exit(1)
		}

		// Repair the old JSON data using the repairdata package
		newJSONBytes, err := repairdata.RepairSessionData(oldJSONBytes)
		if err != nil {
			fmt.Printf("Error repairing the JSON data: %s\n", err)
			os.Exit(1)
		}

		// Write the repaired JSON data to a new file
		newFilePath := strings.TrimSuffix(jsonFilePath, ".json") + "_repaired.json"
		err = os.WriteFile(newFilePath, newJSONBytes, 0644)
		if err != nil {
			fmt.Printf("Error writing the new JSON data to file: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Repaired JSON data has been saved to: %s\n", newFilePath)

		// Update the jsonFilePath to the new repaired file path
		jsonFilePath = newFilePath
		// exit the program after repairing the data
		os.Exit(0)
	}

	store, err := exporter.ReadJSONFromFile(jsonFilePath)
	if err != nil {
		fmt.Printf("Error reading or parsing the JSON file: %s\n", err)
		os.Exit(1)
	}

	// Get the output format option from the user
	outputOption := promptForInput(reader, "Select the output format:\n1) CSV\n2) Hugging Face Dataset\n")
	// Process the output option
	processOutputOption(ctx, reader, outputOption, store.ChatNextWebStore.Sessions)
}

// Context Signal
func setupSignalHandling(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		fmt.Println("Signal received, cancelling operations...")
		cancel()
	}()
}

func promptForInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(input)
}

// processOutputOption processes the chosen output option based on user input.
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

	csvFileName := ""
	if formatOption != 3 {
		csvFileName = promptForInput(reader, "Enter the name of the CSV file to save: ")
	}

	switch formatOption {
	case 3:
		sessionsFileName := promptForInput(reader, "Enter the name of the sessions CSV file to save: ")
		messagesFileName := promptForInput(reader, "Enter the name of the messages CSV file to save: ")

		err := exporter.CreateSeparateCSVFiles(sessions, sessionsFileName, messagesFileName)
		if err != nil {
			fmt.Printf("Failed to create separate CSV files: %s\n", err)
			return
		}

		fmt.Printf("Sessions data saved to %s\n", sessionsFileName)
		fmt.Printf("Messages data saved to %s\n", messagesFileName)
	default:
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
}

// processDatasetOption processes the Hugging Face Dataset output option selected by the user.
func processDatasetOption(reader *bufio.Reader, sessions []exporter.Session) {
	datasetOutput, err := exporter.ExtractToDataset(sessions)
	if err != nil {
		fmt.Printf("Failed to extract to dataset: %s\n", err)
		return
	}
	saveToFile(reader, datasetOutput, "dataset")
}

// saveToFile saves the output content to a file based on user input.
func saveToFile(reader *bufio.Reader, content string, fileType string) {
	saveOutput := promptForInput(reader, fmt.Sprintf("Do you want to save the output to a file? (yes/no)\n"))
	if saveOutput == "yes" {
		fileName := promptForInput(reader, fmt.Sprintf("Enter the name of the %s file to save: ", fileType))
		if fileType == "dataset" {
			fileName += ".json"
		}
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("Failed to create the %s file: %s\n", fileType, err)
			return
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				fmt.Printf("Failed to close the %s file: %s\n", fileType, cerr)
			}
		}()

		_, err = file.WriteString(content)
		if err != nil {
			fmt.Printf("Failed to write to the %s file: %s\n", fileType, err)
			return
		}
		fmt.Printf("%s output saved to %s\n", strings.ToTitle(fileType), fileName)
	}
}
