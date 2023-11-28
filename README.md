# ChatGPT Next Web Session Exporter
[![Go Report Card](https://goreportcard.com/badge/github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter)](https://goreportcard.com/report/github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter)

This repository contains tools that facilitate the conversion of chat session data from JSON format to various CSV formats and Hugging Face datasets. It provides a Bash script and a Go program designed to cater to different requirements for data processing and readability, offering multiple output options.

## Description

The tools process a JSON file containing chat session data, which includes fields such as `id`, `topic`, `memoryPrompt`, and a nested `messages` array with message metadata. Both the Bash script and the Go program offer four distinct output options for CSV formatting:

1. **Inline Formatting**: All messages are included in a single cell for each session.
2. **One Message Per Line**: Each message is placed on a new line with session context repeated.
3. **Separate Files for Sessions and Messages**: Two CSV files are created; one for session metadata and one for messages.
4. **JSON String in CSV**: Messages are stored as a JSON string in a single cell, preserving the array structure.

Additionally, the Go program can convert the sessions into a JSON format suitable for use as a Hugging Face dataset.

## Example Output

Below is an example of what the CSV output might look like for each format option:

### Option 1: Inline Formatting

| id                  | topic           | memoryPrompt | messages                                                                                                        |
|---------------------|-----------------|--------------|-----------------------------------------------------------------------------------------------------------------|
| 8dgQves8ClEy0T4vfHjLs | New Conversation | Example prompt | '[user, 11/27/2023, 10:14:00 AM] "hello"; [assistant, 11/27/2023, 10:14:00 AM] "Hello! How can I assist you today?"' |

### Option 2: One Message Per Line

| session_id           | message_id | date                 | role      | content                            | memoryPrompt   |
|----------------------|------------|----------------------|-----------|------------------------------------|----------------|
| 8dgQves8ClEy0T4vfHjLs | ZKSQGCgGKgrtBCSoqLhFe | 11/27/2023, 10:14:00 AM | user      | hello                              | Example prompt |
| 8dgQves8ClEy0T4vfHjLs | S7DZB9nPoMk4Go_30zESE | 11/27/2023, 10:14:00 AM | assistant | Hello! How can I assist you today? | Example prompt |

### Option 3: Separate Files for Sessions and Messages

**sessions.csv:**

| id                  | topic           | memoryPrompt   | ... |
|---------------------|-----------------|----------------|-----|
| 8dgQves8ClEy0T4vfHjLs | New Conversation | Example prompt | ... |

**messages.csv:**

| session_id           | message_id | date                 | role      | content                            | memoryPrompt   |
|----------------------|------------|----------------------|-----------|------------------------------------|----------------|
| 8dgQves8ClEy0T4vfHjLs | ZKSQGCgGKgrtBCSoqLhFe | 11/27/2023, 10:14:00 AM | user      | hello                              | Example prompt |
| 8dgQves8ClEy0T4vfHjLs | S7DZB9nPoMk4Go_30zESE | 11/27/2023, 10:14:00 AM | assistant | Hello! How can I assist you today? | Example prompt |

### Option 4: JSON String in CSV

| id                  | topic           | memoryPrompt   | messages                                                                                                                                          |
|---------------------|-----------------|----------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| 8dgQves8ClEy0T4vfHjLs | New Conversation | Example prompt | [{"id": "ZKSQGCgGKgrtBCSoqLhFe", "date": "11/27/2023, 10:14:00 AM", "role": "user", "content": "hello"}, {"id": "S7DZB9nPoMk4Go_30zESE", "date": "11/27/2023, 10:14:00 AM", "role": "assistant", "content": "Hello! How can I assist you today?"}] |

Note: "..." represents other columns that would be present in the CSV but are omitted here for brevity.

## Usage

### Bash Script

To use the Bash script, follow these steps:

1. Clone the repository or download the `chat_session_exporter.sh` file.
2. Make the script executable:
   ```bash
   chmod +x chat_session_exporter.sh
   ```
3. Run the script and follow the prompts:
   ```bash
   ./chat_session_exporter.sh
   ```

You will be asked to provide the path to your JSON file and to choose your preferred CSV output format. Optionally, you can save the output to a file.

#### Requirements for Bash Script

- `jq`: The script relies on the `jq` command-line JSON processor. Make sure it is installed on your system.
- Bash shell: The script is intended to be run in a Bash environment.

### Go Program

To use the Go program (`main.go`), follow these steps:

1. Clone the repository or download the `main.go` file.
2. Ensure you have Go installed on your system. You can download it from [the official Go website](https://go.dev/dl/).
3. Navigate to the directory containing `main.go` in a terminal.
4. Compile the program using Go:
   ```bash
   go build -o chat_session_exporter main.go
   ```
5. Run the compiled program and follow the prompts:
   ```bash
   ./chat_session_exporter
   ```

You will be asked to provide the path to your JSON file and to choose your preferred output format. Optionally, you can save the output to a file.

#### Requirements for Go Program

- Go programming language installed on your system.
- A JSON file containing the chat session data.

## Contributing

Contributions to improve the tools or extend their functionality are welcome. Please feel free to fork the repository and submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
