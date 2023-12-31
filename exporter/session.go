// Package exporter provides tools for extracting and converting chat session data
// from JSON files into various formats, such as CSV and JSON datasets.
//
// This package facilitates tasks like data visualization, reporting, and machine learning data preparation.
//
// The exporter package defines types to represent chat sessions, messages, and associated metadata.
//
// It includes functions to:
//
//   - Read chat session data from JSON files
//   - Convert sessions to CSV with different formatting options
//   - Create separate CSV files for sessions and messages
//   - Extract sessions to a JSON format for Hugging Face datasets
//
// The package also handles fields in the source JSON that may be represented as either
// strings or integers by using the custom StringOrInt type.
//
// Additionally, it now supports context-aware operations, allowing for better control
// over long-running processes and the ability to cancel them if needed.
//
// Code:
//
//	func (soi *StringOrInt) UnmarshalJSON(data []byte) error {
//		// Try unmarshalling into a string
//		var s string
//		if err := json.Unmarshal(data, &s); err != nil {
//			// If there is an error, try unmarshalling into an int
//			var i int64
//			if err := json.Unmarshal(data, &i); err != nil {
//				return err // Return the error if it is not a string or int
//			}
//			// Convert int to string and assign it to the custom type
//			*soi = StringOrInt(strconv.FormatInt(i, 10))
//			return nil
//		}
//		// If no error, assign the string value to the custom type
//		*soi = StringOrInt(s)
//		return nil
//	}
//
// Usage examples:
//
// To read chat sessions from a JSON file and convert them to a CSV format with context support:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	store, err := exporter.ReadJSONFromFile("path/to/chat-sessions.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	err = exporter.ConvertSessionsToCSV(ctx, store.ChatNextWebStore.Sessions, exporter.FormatOptionInline, "output.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// To create separate CSV files for sessions and messages:
//
//	err = exporter.CreateSeparateCSVFiles(store.ChatNextWebStore.Sessions, "sessions.csv", "messages.csv")
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
// Copyright (c) 2023 H0llyW00dzZ
package exporter

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// FormatOptionInline specifies the format where messages are displayed inline.
	FormatOptionInline = iota + 1

	// FormatOptionPerLine specifies the format where each message is on a separate line.
	FormatOptionPerLine

	// FormatOptionJSON specifies the format where messages are encoded as JSON.
	FormatOptionJSON

	// OutputFormatSeparateCSVFiles specifies the option to create separate CSV files for sessions and messages.
	OutputFormatSeparateCSVFiles
)

// StringOrInt is a custom type to handle JSON values that can be either strings or integers (Magic Golang 🎩 🪄).
//
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

// ReadJSONFromFile reads a JSON file from the given file path and unmarshals it into a ChatNextWebStore struct.
//
// It returns an error if the file cannot be opened, the JSON
// is invalid, or the JSON format does not match the expected ChatNextWebStore format.
func ReadJSONFromFile(filePath string) (ChatNextWebStore, error) {
	// Variable `store` is of type ChatNextWebStore. It is used to store the unmarshaled JSON data.
	var store ChatNextWebStore

	// Variable `file` is of type *os.File. It holds the pointer to the opened JSON file.
	// Variable `err` is of type error. It is used to capture any errors that occur during the file opening and JSON decoding process.
	file, err := os.Open(filePath)
	if err != nil {
		// If an error occurs while opening the file, the function returns the empty `store` and the error.
		return store, err
	}
	// Defer the closing of the file until the function exits.
	// This ensures that the file is closed properly to free resources and avoid leaks.
	defer file.Close()

	// Variable `decoder` is of type *json.Decoder. It is used to decode the JSON file into the `store` struct.
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&store)
	if err != nil {
		// If an error occurs during decoding, the function returns the empty `store` and the error.
		return store, err
	}

	// Check if the `Sessions` field in `store.ChatNextWebStore` is nil, which indicates the JSON was not in the expected format.
	if store.ChatNextWebStore.Sessions == nil {
		// If the JSON format is incorrect, the function returns the empty `store` and a format error.
		return store, fmt.Errorf("JSON does not match the expected format chat-next-web-store")
	}

	// If no error occurs, the function returns the populated `store` and a nil error.
	return store, nil
}

// ConvertSessionsToCSV writes a slice of Session objects into a CSV file with support for context cancellation.
//
// It delegates the writing of sessions to format-specific functions based on the formatOption provided.
//
// The outputFilePath parameter specifies the path to the output CSV file.
//
// It returns an error if the context is cancelled, the format option is invalid, or writing to the CSV fails.
func ConvertSessionsToCSV(ctx context.Context, sessions []Session, formatOption int, outputFilePath string) error {
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output CSV file: %w", err)
	}
	defer outputFile.Close()

	csvWriter := csv.NewWriter(outputFile)
	defer csvWriter.Flush()

	headers, err := getCSVHeaders(formatOption)
	if err != nil {
		return err
	}

	if err := WriteHeaders(csvWriter, headers); err != nil {
		return err
	}

	writeFunc, err := getWriteFunction(formatOption)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if err := checkContextCancellation(ctx); err != nil {
			return err
		}

		if err := writeFunc(csvWriter, session); err != nil {
			return err
		}
	}

	return nil
}

// getCSVHeaders returns the headers for the CSV file based on the formatOption.
// It returns an error if the formatOption is not recognized.
func getCSVHeaders(formatOption int) ([]string, error) {
	switch formatOption {
	case FormatOptionInline:
		return []string{"id", "topic", "memoryPrompt", "messages"}, nil
	case FormatOptionPerLine:
		return []string{"session_id", "message_id", "date", "role", "content", "memoryPrompt"}, nil
	case FormatOptionJSON:
		return []string{"id", "topic", "memoryPrompt", "messages"}, nil
	default:
		return nil, fmt.Errorf("invalid format option")
	}
}

// getWriteFunction returns a function that corresponds to the CSV writing strategy for the given formatOption.
// The returned function takes a csv.Writer and a Session object to write the session data according to the format.
// It returns an error if the formatOption is not recognized.
func getWriteFunction(formatOption int) (func(*csv.Writer, Session) error, error) {
	switch formatOption {
	case FormatOptionInline:
		return writeInlineFormat, nil
	case FormatOptionPerLine:
		return writePerLineFormat, nil
	case FormatOptionJSON:
		return writeJSONFormat, nil
	default:
		return nil, fmt.Errorf("invalid format option")
	}
}

// writeInlineFormat writes session data in an inline format to the provided csv.Writer.
// Messages are concatenated into a single string with a delimiter.
// It returns an error if writing to the CSV fails.
func writeInlineFormat(csvWriter *csv.Writer, session Session) error {
	var messageContents []string
	for _, message := range session.Messages {
		messageContents = append(messageContents, fmt.Sprintf("[%s, %s] \"%s\"", message.Role, message.Date, message.Content))
	}
	sessionData := []string{session.ID, session.Topic, session.MemoryPrompt, strings.Join(messageContents, "; ")}
	return csvWriter.Write(sessionData)
}

// writePerLineFormat writes each message of a session on a new line in the provided csv.Writer.
// It returns an error if writing to the CSV fails.
func writePerLineFormat(csvWriter *csv.Writer, session Session) error {
	for _, message := range session.Messages {
		sessionData := []string{session.ID, message.ID, message.Date, message.Role, message.Content, session.MemoryPrompt}
		if err := csvWriter.Write(sessionData); err != nil {
			return err
		}
	}
	return nil
}

// writeJSONFormat writes session data with messages as a JSON string to the provided csv.Writer.
// It returns an error if marshaling messages to JSON or writing to the CSV fails.
func writeJSONFormat(csvWriter *csv.Writer, session Session) error {
	messagesJSON, err := json.Marshal(session.Messages)
	if err != nil {
		return err
	}
	sessionData := []string{session.ID, session.Topic, session.MemoryPrompt, string(messagesJSON)}
	return csvWriter.Write(sessionData)
}

// checkContextCancellation checks if the context has been cancelled.
// It returns a non-nil error if the context is cancelled; otherwise, it returns nil.
func checkContextCancellation(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// WriteHeaders writes the provided headers to the csv.Writer.
func WriteHeaders(csvWriter *csv.Writer, headers []string) error {
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}
	return nil
}

// WriteSessionData writes session data to the provided csv.Writer.
func WriteSessionData(csvWriter *csv.Writer, sessions []Session) error {
	for _, session := range sessions {
		sessionData := []string{
			session.ID, session.Topic, session.MemoryPrompt,
		}
		if err := csvWriter.Write(sessionData); err != nil {
			return fmt.Errorf("failed to write session data: %w", err)
		}
	}
	return nil
}

// WriteMessageData writes message data to the provided csv.Writer.
func WriteMessageData(csvWriter *csv.Writer, sessions []Session) error {
	for _, session := range sessions {
		for _, message := range session.Messages {
			messageData := []string{
				session.ID, message.ID, message.Date, message.Role, message.Content, session.MemoryPrompt,
			}
			if err := csvWriter.Write(messageData); err != nil {
				return fmt.Errorf("failed to write message data: %w", err)
			}
		}
	}
	return nil
}

// initializeCSVFile creates and initializes a CSV file with the given name and headers.
func initializeCSVFile(fileName string, headers []string) (*os.File, *csv.Writer, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create file %s: %w", fileName, err)
	}

	csvWriter := csv.NewWriter(file)

	if err := WriteHeaders(csvWriter, headers); err != nil {
		file.Close() // ignore error; we're already handling an error
		return nil, nil, err
	}

	return file, csvWriter, nil
}

// closeCSVWriter closes the csv.Writer and the underlying file, and checks for errors.
func closeCSVWriter(csvWriter *csv.Writer, file *os.File) error {
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		file.Close() // ignore error; we're already handling an error
		return fmt.Errorf("failed to flush data: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// CreateSeparateCSVFiles creates two separate CSV files for sessions and messages from a slice of Session objects.
//
// It takes the file names as parameters and returns an error if the files cannot be created or if writing the data fails.
//
// Errors from closing files or flushing data to the CSV writers are captured and will be returned after all operations are attempted.
//
// Error messages are logged to the console.
func CreateSeparateCSVFiles(sessions []Session, sessionsFileName string, messagesFileName string) (err error) {
	// Create and initialize the sessions CSV file.
	var sessionsFile *os.File
	var sessionsWriter *csv.Writer
	sessionsFile, sessionsWriter, err = initializeCSVFile(sessionsFileName, []string{"id", "topic", "memoryPrompt"})
	if err != nil {
		return err
	}
	defer func() {
		if cerr := closeCSVWriter(sessionsWriter, sessionsFile); cerr != nil {
			err = cerr
		}
	}()

	// Write session data.
	if err = WriteSessionData(sessionsWriter, sessions); err != nil {
		return err
	}

	// Create and initialize the messages CSV file.
	var messagesFile *os.File
	var messagesWriter *csv.Writer
	messagesFile, messagesWriter, err = initializeCSVFile(messagesFileName, []string{"session_id", "message_id", "date", "role", "content", "memoryPrompt"})
	if err != nil {
		return err
	}
	defer func() {
		if cerr := closeCSVWriter(messagesWriter, messagesFile); cerr != nil {
			err = cerr
		}
	}()

	// Write message data.
	if err = WriteMessageData(messagesWriter, sessions); err != nil {
		return err
	}

	return nil
}

// ExtractToDataset converts a slice of Session objects into a JSON formatted string suitable for use as a dataset in machine learning applications.
//
// It returns an error if marshaling the sessions into JSON format fails.
func ExtractToDataset(sessions []Session) (string, error) {
	dataset := make(map[string][]Session)
	dataset["dataset"] = sessions

	jsonData, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
