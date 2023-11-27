#!/bin/bash

# Function to convert JSON sessions to CSV
convert_sessions_to_csv() {
    local format_option=$1
    case $format_option in
        1) # Inline Formatting
            jq -r '.["chat-next-web-store"].sessions[] | [ .id, .topic, .memoryPrompt, .stat.tokenCount, .stat.wordCount, .stat.charCount, .lastUpdate, .lastSummarizeIndex, .mask.id, .mask.avatar, .mask.name, .mask.lang, .mask.createdAt, (.messages | map("[\(.role), \(.date)] \(.content|tostring)") | join("; ")) ] | @csv'
            ;;
        2) # One Message Per Line
            jq -r '.["chat-next-web-store"].sessions[] | .id as $session_id | .memoryPrompt as $memoryPrompt | .messages[] | [$session_id, .id, .date, .role, .content, $memoryPrompt] | @csv'
            ;;
        3) # Separate Files for Sessions and Messages (sessions part)
            jq -r '.["chat-next-web-store"].sessions[] | [ .id, .topic, .memoryPrompt, .stat.tokenCount, .stat.wordCount, .stat.charCount, .lastUpdate, .lastSummarizeIndex, .mask.id, .mask.avatar, .mask.name, .mask.lang, .mask.createdAt] | @csv'
            ;;
        4) # JSON String in CSV
            jq -r '.["chat-next-web-store"].sessions[] | [ .id, .topic, .memoryPrompt, .stat.tokenCount, .stat.wordCount, .stat.charCount, .lastUpdate, .lastSummarizeIndex, .mask.id, .mask.avatar, .mask.name, .mask.lang, .mask.createdAt, (.messages | @json) ] | @csv'
            ;;
        *)
            echo "Invalid format option."
            exit 1
            ;;
    esac
}

# Prompt the user for the JSON file path
echo "Enter the path to the JSON file:"
read -r json_file_path

# Check if the JSON file exists
if [[ ! -f "$json_file_path" ]]; then
    echo "The file $json_file_path does not exist."
    exit 1
fi

# Prompt the user for the output option
echo "Select the message output format:"
echo "1) Inline Formatting"
echo "2) One Message Per Line"
echo "3) Separate Files for Sessions and Messages"
echo "4) JSON String in CSV"
read -r format_option

# Read the JSON content
json_content=$(<"$json_file_path")

# Convert the sessions to CSV according to the chosen format
csv_output=$(echo "$json_content" | convert_sessions_to_csv "$format_option")

# Output the CSV content
echo "$csv_output"

# If option 3 is chosen, handle messages separately
if [[ $format_option == "3" ]]; then
    messages_output=$(jq -r '.["chat-next-web-store"].sessions[] | .id as $session_id | .memoryPrompt as $memoryPrompt | .messages[] | [$session_id, .id, .date, .role, .content, $memoryPrompt] | @csv' <<< "$json_content")
    echo "$messages_output"
fi

# Optionally, you can save this output to a file
echo "Do you want to save the output to a file? (yes/no)"
read -r save_output

if [[ $save_output == "yes" ]]; then
    echo "Enter the name of the CSV file to save:"
    read -r csv_file_name
    if [[ $format_option == "3" ]]; then
        echo "$csv_output" > "${csv_file_name}_sessions.csv"
        echo "$messages_output" > "${csv_file_name}_messages.csv"
        echo "Session output saved to ${csv_file_name}_sessions.csv"
        echo "Messages output saved to ${csv_file_name}_messages.csv"
    else
        echo "$csv_output" > "$csv_file_name"
        echo "Output saved to $csv_file_name"
    fi
fi
