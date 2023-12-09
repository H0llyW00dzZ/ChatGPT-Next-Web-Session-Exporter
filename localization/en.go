// Go localization CLI provides support for internationalization.
//
// Note: WIP Not ready for use.
//
// Copyright (c) 2023 H0llyW00dzZ
package localization

type Locale string

const (
	English Locale = "en"
	// Add other locales here
)

// Messages contains the localized strings for a given locale.
type Messages map[string]string

// locales holds the localization data for each supported locale.
var locales = map[Locale]Messages{
	English: {
		"FileTypeDataset":                "dataset",
		"PromptEnterJSONFilePath":        "Enter the path to the JSON file: ",
		"PromptRepairData":               "Do you want to repair data? (yes/no): ",
		"PromptSelectOutputFormat":       "Select the output format:\n1) CSV\n2) Hugging Face Dataset\n",
		"PromptSelectCSVOutputFormat":    "Select the message output format:\n1) Inline Formatting\n2) One Message Per Line\n3) Separate Files for Sessions and Messages\n4) JSON String in CSV\n",
		"PromptEnterCSVFileName":         "Enter the name of the CSV file to save: ",
		"PromptEnterSessionsCSVFileName": "Enter the name of the sessions CSV file to save: ",
		"PromptEnterMessagesCSVFileName": "Enter the name of the messages CSV file to save: ",
		"PromptSaveOutputToFile":         "Do you want to save the output to a file? (yes/no)\n",
		"PromptEnterFileName":            "Enter the name of the %s file to save: ",
		"fileType":                       "dataset",
		"GopherHelper":                   "[GopherHelper] Exiting gracefully...",
		"GopherHelperReason":             "Reason: Operation canceled or end of input. Exiting program.",
		"promptForInput":                 "Enter the path to the JSON file: ",
	},
	// Add other locales with their translations here
}

// Get returns the localized message for the given key and locale.
func Get(locale Locale, key string) string {
	if messages, ok := locales[locale]; ok {
		if msg, ok := messages[key]; ok {
			return msg
		}
	}
	// Fallback to English if the key or locale is not found
	return locales[English][key]
}
