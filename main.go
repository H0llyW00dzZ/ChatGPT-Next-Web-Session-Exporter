// Package main is the entry point for the CLI tool that utilizes the exporter package
// to process chat session data and provide various output formats.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/exporter"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Get the JSON file path from the user
	jsonFilePath := promptForInput(reader, "Enter the path to the JSON file: ")

	// Read the JSON content using the exporter package
	store, err := exporter.ReadJSONFromFile(jsonFilePath)
	if err != nil {
		fmt.Printf("Error reading or parsing the JSON file: %s\n", err)
		os.Exit(1)
	}

	// Get the output format option from the user
	outputOption := promptForInput(reader, "Select the output format:\n1) CSV\n2) Hugging Face Dataset\n")

	// Process the output option
	processOutputOption(reader, outputOption, store.ChatNextWebStore.Sessions)
}

// promptForInput prompts the user for input and returns the trimmed response.
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
func processOutputOption(reader *bufio.Reader, outputOption string, sessions []exporter.Session) {
	switch outputOption {
	case "1":
		processCSVOption(reader, sessions)
	case "2":
		processDatasetOption(reader, sessions)
	default:
		fmt.Println("Invalid output option.")
	}
}

// processCSVOption processes the CSV output option selected by the user.
func processCSVOption(reader *bufio.Reader, sessions []exporter.Session) {
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
		err := exporter.ConvertSessionsToCSV(sessions, formatOption, csvFileName)
		if err != nil {
			fmt.Printf("Failed to convert sessions to CSV: %s\n", err)
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
		fmt.Printf("%s output saved to %s\n", strings.Title(fileType), fileName)
	}
}
