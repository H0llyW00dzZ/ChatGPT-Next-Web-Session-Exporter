package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// Custom type to handle both string and numeric JSON values for ID (Magic Golang 🎩 🪄)
type StringOrInt string

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

// Define Go structs to match the JSON structure
type Message struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Stat struct {
	TokenCount int `json:"tokenCount"`
	WordCount  int `json:"wordCount"`
	CharCount  int `json:"charCount"`
}

type Mask struct {
	ID        StringOrInt `json:"id"` // Use the custom type for ID
	Avatar    string      `json:"avatar"`
	Name      string      `json:"name"`
	Lang      string      `json:"lang"`
	CreatedAt int64       `json:"createdAt"` // Assuming it's a Unix timestamp
}

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

type Store struct {
	Sessions []Session `json:"sessions"`
}

type ChatNextWebStore struct {
	ChatNextWebStore Store `json:"chat-next-web-store"`
}

// Function to read JSON from file
func readJSONFromFile(filePath string) (ChatNextWebStore, error) {
	var store ChatNextWebStore

	file, err := os.Open(filePath)
	if err != nil {
		return store, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&store)
	return store, err
}

func printLine(widths []int) {
	for _, w := range widths {
		fmt.Printf("+-%s-", strings.Repeat("-", w))
	}
	fmt.Println("+")
}

func printTable(headers []string, data [][]string, numRows int) {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range data {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	printLine(widths)
	for i, header := range headers {
		fmt.Printf("| %-*s ", widths[i], header)
	}
	fmt.Println("|")
	printLine(widths)

	for rowIndex, row := range data {
		if numRows >= 0 && rowIndex >= numRows {
			break
		}
		for i, cell := range row {
			fmt.Printf("| %-*s ", widths[i], cell)
		}
		fmt.Println("|")
	}
	printLine(widths)
}

func convertSessionsToCSV(sessions []Session, formatOption int) (string, error) {
	var (
		csvData [][]string
		headers []string
	)

	switch formatOption {
	case 1: // Inline Formatting
		headers = []string{"id", "topic", "memoryPrompt", "messages"}
		for _, session := range sessions {
			var messageContents []string
			for _, message := range session.Messages {
				messageContents = append(messageContents, fmt.Sprintf("[%s, %s] \"%s\"", message.Role, message.Date, message.Content))
			}
			sessionData := []string{
				session.ID, session.Topic, session.MemoryPrompt, strings.Join(messageContents, "; "),
			}
			csvData = append(csvData, sessionData)
		}

	case 2: // One Message Per Line
		headers = []string{"session_id", "message_id", "date", "role", "content", "memoryPrompt"}
		for _, session := range sessions {
			for _, message := range session.Messages {
				sessionData := []string{
					session.ID, message.ID, message.Date, message.Role, message.Content, session.MemoryPrompt,
				}
				csvData = append(csvData, sessionData)
			}
		}

	case 4: // JSON String in CSV
		headers = []string{"id", "topic", "memoryPrompt", "messages"}
		for _, session := range sessions {
			messagesJSON, err := json.Marshal(session.Messages)
			if err != nil {
				return "", err
			}
			sessionData := []string{
				session.ID, session.Topic, session.MemoryPrompt, string(messagesJSON),
			}
			csvData = append(csvData, sessionData)
		}

	default:
		return "", fmt.Errorf("invalid format option")
	}

	// Ask the user if they want to display the table
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to display the table? (yes/no): ")
	displayTable, _ := reader.ReadString('\n')
	displayTable = strings.TrimSpace(displayTable)

	if displayTable == "yes" {
		// Ask the user how many messages to display
		fmt.Print("How many messages do you want to display in the table? (Enter a number): ")
		numMessagesStr, _ := reader.ReadString('\n')
		numMessagesStr = strings.TrimSpace(numMessagesStr)
		numMessages, err := strconv.Atoi(numMessagesStr)
		if err != nil {
			fmt.Println("Invalid number. Displaying all messages.")
			numMessages = -1 // Use -1 as an indicator to display all messages
		}

		// Using tablewriter to print the table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(headers)
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		table.SetCenterSeparator("|")

		// If numMessages is -1, append all rows; otherwise, append up to numMessages rows
		if numMessages < 0 || numMessages > len(csvData) {
			numMessages = len(csvData)
		}

		for _, v := range csvData[:numMessages] {
			table.Append(v)
		}
		table.Render() // Send output
	}

	csvString := &strings.Builder{}
	csvWriter := csv.NewWriter(csvString)
	if err := csvWriter.Write(headers); err != nil {
		return "", err
	}
	if err := csvWriter.WriteAll(csvData); err != nil {
		return "", err
	}
	csvWriter.Flush()

	return csvString.String(), nil
}

// Function to create separate CSV files for sessions and messages
func createSeparateCSVFiles(sessions []Session, sessionsFileName string, messagesFileName string) error {
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

// Function to extract JSON sessions to Hugging Face dataset
func extractToDataset(sessions []Session) (string, error) {
	dataset := make(map[string][]Session)
	dataset["dataset"] = sessions

	jsonData, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Prompt the user for the JSON file path
	fmt.Print("Enter the path to the JSON file: ")
	jsonFilePath, _ := reader.ReadString('\n')
	jsonFilePath = strings.TrimSpace(jsonFilePath)

	// Read the JSON content
	store, err := readJSONFromFile(jsonFilePath)
	if err != nil {
		fmt.Printf("Failed to read the JSON file: %s\n", err)
		return
	}

	// Prompt the user for the output option
	fmt.Println("Select the output format:")
	fmt.Println("1) CSV")
	fmt.Println("2) Hugging Face Dataset")
	outputOption, _ := reader.ReadString('\n')
	outputOption = strings.TrimSpace(outputOption)

	switch outputOption {
	case "1":
		// Prompt the user for the CSV format option
		fmt.Println("Select the message output format:")
		fmt.Println("1) Inline Formatting")
		fmt.Println("2) One Message Per Line")
		fmt.Println("3) Separate Files for Sessions and Messages")
		fmt.Println("4) JSON String in CSV")
		formatOptionStr, _ := reader.ReadString('\n')
		formatOption, err := strconv.Atoi(strings.TrimSpace(formatOptionStr))
		if err != nil {
			fmt.Println("Invalid format option.")
			return
		}

		if formatOption == 3 {
			fmt.Println("Enter the name of the sessions CSV file to save:")
			sessionsFileName, _ := reader.ReadString('\n')
			sessionsFileName = strings.TrimSpace(sessionsFileName)

			fmt.Println("Enter the name of the messages CSV file to save:")
			messagesFileName, _ := reader.ReadString('\n')
			messagesFileName = strings.TrimSpace(messagesFileName)

			err := createSeparateCSVFiles(store.ChatNextWebStore.Sessions, sessionsFileName, messagesFileName)
			if err != nil {
				fmt.Printf("Failed to create separate CSV files: %s\n", err)
				return
			}

			fmt.Printf("Sessions data saved to %s\n", sessionsFileName)
			fmt.Printf("Messages data saved to %s\n", messagesFileName)
		} else {
			// Convert the sessions to CSV according to the chosen format
			csvOutput, err := convertSessionsToCSV(store.ChatNextWebStore.Sessions, formatOption)
			if err != nil {
				fmt.Printf("Failed to convert sessions to CSV: %s\n", err)
				return
			}

			// Output the CSV content
			//fmt.Println(csvOutput)

			// Optionally, you can save this output to a file
			fmt.Println("Do you want to save the output to a file? (yes/no)")
			saveOutput, _ := reader.ReadString('\n')
			saveOutput = strings.TrimSpace(saveOutput)
			if saveOutput == "yes" {
				fmt.Println("Enter the name of the CSV file to save:")
				csvFileName, _ := reader.ReadString('\n')
				csvFileName = strings.TrimSpace(csvFileName)
				file, err := os.Create(csvFileName)
				if err != nil {
					fmt.Printf("Failed to create the CSV file: %s\n", err)
					return
				}
				defer file.Close()
				_, err = file.WriteString(csvOutput)
				if err != nil {
					fmt.Printf("Failed to write to the CSV file: %s\n", err)
					return
				}
				fmt.Printf("CSV output saved to %s\n", csvFileName)
			}
		}

	case "2":
		// Extract the JSON sessions to Hugging Face dataset format
		datasetOutput, err := extractToDataset(store.ChatNextWebStore.Sessions)
		if err != nil {
			fmt.Printf("Failed to extract to dataset: %s\n", err)
			return
		}

		// Output the dataset content
		fmt.Println(datasetOutput)

		// Optionally, you can save this output to a file
		fmt.Println("Do you want to save the output to a file? (yes/no)")
		saveOutput, _ := reader.ReadString('\n')
		saveOutput = strings.TrimSpace(saveOutput)
		if saveOutput == "yes" {
			fmt.Println("Enter the name of the dataset file to save (without the extension):")
			datasetFileName, _ := reader.ReadString('\n')
			datasetFileName = strings.TrimSpace(datasetFileName)
			file, err := os.Create(datasetFileName + ".json")
			if err != nil {
				fmt.Printf("Failed to create the dataset file: %s\n", err)
				return
			}
			defer file.Close()
			_, err = file.WriteString(datasetOutput)
			if err != nil {
				fmt.Printf("Failed to write to the dataset file: %s\n", err)
				return
			}
			fmt.Printf("Dataset output saved to %s.json\n", datasetFileName)
		}

	default:
		fmt.Println("Invalid output option.")
	}
}
