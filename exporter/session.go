// Package exporter provides tools for extracting chat session data from JSON files
// and converting it into various formats such as CSV and JSON datasets. This package
// is designed to facilitate the analysis and processing of chat data, making it easier
// to perform tasks such as data visualization, reporting, or feeding the data into
// machine learning models.
//
// The exporter package defines several types to represent chat sessions, messages,
// and associated metadata. It also includes functions to read chat session data from
// JSON files, convert sessions to CSV with different formatting options, create separate
// CSV files for sessions and messages, and extract sessions to a JSON format suitable
// for use with Hugging Face datasets.
//
// Usage:
//
// To read chat sessions from a JSON file and convert them to a CSV format:
//
//	store, err := exporter.ReadJSONFromFile("path/to/chat-sessions.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	csvData, err := exporter.ConvertSessionsToCSV(store.ChatNextWebStore.Sessions, exporter.FormatOptionInline, "output.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(csvData)
//
// To create separate CSV files for sessions and messages:
//
//	err := exporter.CreateSeparateCSVFiles(store.ChatNextWebStore.Sessions, "sessions.csv", "messages.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// To extract chat sessions to a JSON dataset:
//
//	datasetJSON, err := exporter.ExtractToDataset(store.ChatNextWebStore.Sessions)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(datasetJSON)
//
// The package supports handling of IDs and other fields that may be represented as either
// strings or integers in the source JSON by using the custom StringOrInt type.
package exporter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// StringOrInt is a custom type to handle JSON values that can be either strings or integers (Magic Golang 🎩 🪄).
// It implements the Unmarshaler interface to handle this mixed type when unmarshaling JSON data.
type StringOrInt string

// UnmarshalJSON is a custom unmarshaler for StringOrInt that tries to unmarshal JSON data
// as a string, and if that fails, as an integer, which is then converted to a string.
func (soi *StringOrInt) UnmarshalJSON(data []byte) error {
	// Try unmarshalling into a string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// If there is an error, try unmarshalling into an int
		var i int64
		if err := json.Unmarshal(data, &i); err != nil {
			return err // Return the error if it is not a string or int
		}
		// Convert int to string and assign it to the custom type
		*soi = StringOrInt(strconv.FormatInt(i, 10))
		return nil
	}
	// If no error, assign the string value to the custom type
	*soi = StringOrInt(s)
	return nil
}

// Message represents a single message within a chat session, including metadata
// like the ID, date, role of the sender, and the content of the message itself.
type Message struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Stat represents statistics for a chat session, such as the count of tokens,
// words, and characters.
type Stat struct {
	TokenCount int `json:"tokenCount"`
	WordCount  int `json:"wordCount"`
	CharCount  int `json:"charCount"`
}

// Mask represents an anonymization mask for a participant in a chat session,
// including the participant's ID, avatar link, name, language, and creation timestamp.
type Mask struct {
	ID        StringOrInt `json:"id"` // Use the custom type for ID
	Avatar    string      `json:"avatar"`
	Name      string      `json:"name"`
	Lang      string      `json:"lang"`
	CreatedAt int64       `json:"createdAt"` // Assuming it's a Unix timestamp
}

// Session represents a single chat session, including session metadata,
// statistics, messages, and the mask for the participant.
type Session struct {
	ID                 string    `json:"id"`
	Topic              string    `json:"topic"`
	MemoryPrompt       string    `json:"memoryPrompt"`
	Stat               Stat      `json:"stat"`
	LastUpdate         int64     `json:"lastUpdate"` // Changed to int64 assuming it's a Unix timestamp
	LastSummarizeIndex int       `json:"lastSummarizeIndex"`
	Mask               Mask      `json:"mask"`
	Messages           []Message `json:"messages"`
}

// Store encapsulates a collection of chat sessions.
type Store struct {
	Sessions []Session `json:"sessions"`
}

// ChatNextWebStore is a wrapper for Store that aligns with the expected JSON structure
// for a chat-next-web-store object.
type ChatNextWebStore struct {
	ChatNextWebStore Store `json:"chat-next-web-store"`
}

// ReadJSONFromFile reads a JSON file from the given file path and unmarshals it into
// a ChatNextWebStore struct. It returns an error if the file cannot be opened, the JSON
// is invalid, or the JSON format does not match the expected ChatNextWebStore format.
func ReadJSONFromFile(filePath string) (ChatNextWebStore, error) {
	var store ChatNextWebStore

	file, err := os.Open(filePath)
	if err != nil {
		return store, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&store)
	if err != nil {
		return store, err
	}

	// Check if the store.ChatNextWebStore.Sessions is nil, which indicates the JSON was not in the expected format.
	if store.ChatNextWebStore.Sessions == nil {
		return store, fmt.Errorf("JSON does not match the expected format chat-next-web-store")
	}

	return store, nil
}

// ConvertSessionsToCSV writes a slice of Session objects into a CSV file.
// It formats the CSV data in different ways based on the formatOption parameter.
// It returns an error if the format option is invalid or if writing the CSV data fails.
func ConvertSessionsToCSV(sessions []Session, formatOption int, outputFilePath string) error {
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output CSV file: %w", err)
	}
	defer outputFile.Close()

	csvWriter := csv.NewWriter(outputFile)
	defer csvWriter.Flush()

	// Define headers based on the formatOption
	var headers []string
	switch formatOption {
	case 1: // Inline Formatting
		headers = []string{"id", "topic", "memoryPrompt", "messages"}
	case 2: // One Message Per Line
		headers = []string{"session_id", "message_id", "date", "role", "content", "memoryPrompt"}
	case 4: // JSON String in CSV
		headers = []string{"id", "topic", "memoryPrompt", "messages"}
	default:
		return fmt.Errorf("invalid format option")
	}

	// Write the headers to the CSV
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers to CSV: %w", err)
	}

	// Write each session to the CSV based on the selected format
	for _, session := range sessions {
		var sessionData []string
		switch formatOption {
		case 1: // Inline Formatting
			var messageContents []string
			for _, message := range session.Messages {
				messageContents = append(messageContents, fmt.Sprintf("[%s, %s] \"%s\"", message.Role, message.Date, message.Content))
			}
			sessionData = []string{session.ID, session.Topic, session.MemoryPrompt, strings.Join(messageContents, "; ")}
		case 2: // One Message Per Line
			for _, message := range session.Messages {
				sessionData = []string{session.ID, message.ID, message.Date, message.Role, message.Content, session.MemoryPrompt}
				if err := csvWriter.Write(sessionData); err != nil {
					return fmt.Errorf("failed to write session data to CSV: %w", err)
				}
			}
			continue // Skip the default write for this format option
		case 4: // JSON String in CSV
			messagesJSON, err := json.Marshal(session.Messages)
			if err != nil {
				return fmt.Errorf("failed to marshal messages to JSON: %w", err)
			}
			sessionData = []string{session.ID, session.Topic, session.MemoryPrompt, string(messagesJSON)}
		}
		// Write the session data to the CSV
		if err := csvWriter.Write(sessionData); err != nil {
			return fmt.Errorf("failed to write session data to CSV: %w", err)
		}
	}

	return nil
}

// CreateSeparateCSVFiles creates two separate CSV files for sessions and messages from
// a slice of Session objects. It takes the file names as parameters and returns an error
// if the files cannot be created or if writing the data fails.
func CreateSeparateCSVFiles(sessions []Session, sessionsFileName string, messagesFileName string) error {
	// Create CSV file for sessions
	sessionsFile, err := os.Create(sessionsFileName)
	if err != nil {
		return err
	}
	defer sessionsFile.Close()

	sessionsWriter := csv.NewWriter(sessionsFile)
	defer sessionsWriter.Flush()

	sessionsHeaders := []string{"id", "topic", "memoryPrompt"} // Add other headers as needed
	if err := sessionsWriter.Write(sessionsHeaders); err != nil {
		return err
	}
	for _, session := range sessions {
		sessionData := []string{
			session.ID, session.Topic, session.MemoryPrompt,
		}
		if err := sessionsWriter.Write(sessionData); err != nil {
			return err
		}
	}

	// Create CSV file for messages
	messagesFile, err := os.Create(messagesFileName)
	if err != nil {
		return err
	}
	defer messagesFile.Close()

	messagesWriter := csv.NewWriter(messagesFile)
	defer messagesWriter.Flush()

	messagesHeaders := []string{"session_id", "message_id", "date", "role", "content", "memoryPrompt"}
	if err := messagesWriter.Write(messagesHeaders); err != nil {
		return err
	}
	for _, session := range sessions {
		for _, message := range session.Messages {
			messageData := []string{
				session.ID, message.ID, message.Date, message.Role, message.Content, session.MemoryPrompt,
			}
			if err := messagesWriter.Write(messageData); err != nil {
				return err
			}
		}
	}

	return nil
}

// ExtractToDataset converts a slice of Session objects into a JSON formatted string
// suitable for use as a dataset in machine learning applications. It returns an error
// if marshaling the sessions into JSON format fails.
func ExtractToDataset(sessions []Session) (string, error) {
	dataset := make(map[string][]Session)
	dataset["dataset"] = sessions

	jsonData, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
