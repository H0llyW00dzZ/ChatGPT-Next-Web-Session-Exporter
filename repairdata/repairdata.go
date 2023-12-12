// Package repairdata provides utilities for transforming JSON data from an old format to a new format.
//
// It specifically ensures that each session's modelConfig contains a 'systemprompt' field.
//
// Copyright (c) 2023 H0llyW00dzZ
package repairdata

import (
	"encoding/json"
	"strconv"
	"time"
)

// StringOrInt is a custom type to handle JSON values that can be either strings or integers (Magic Golang 🎩 🪄).
//
// It implements the Unmarshaler interface to handle this mixed type when unmarshaling JSON data.
type StringOrInt string

// UnmarshalJSON is a custom unmarshaler for StringOrInt that tries to unmarshal JSON data
//
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

// OldData represents the structure of the old JSON data format.
//
// Additional fields can be added to match the complete old JSON structure as needed.
type OldData struct {
	ChatNextWebStore struct {
		Sessions            []Session `json:"sessions"`
		CurrentSessionIndex int       `json:"currentSessionIndex"`
		LastUpdateTime      int64     `json:"lastUpdateTime"`
	} `json:"chat-next-web-store"`
	// Include other fields from the old JSON format...
}

// NewData represents the structure of the new JSON data format.
//
// Additional fields can be added to match the complete new JSON structure as needed.
type NewData struct {
	ChatNextWebStore struct {
		Sessions            []Session `json:"sessions"`
		CurrentSessionIndex int       `json:"currentSessionIndex"`
		LastUpdateTime      int64     `json:"lastUpdateTime"`
	} `json:"chat-next-web-store"`
	// Include other fields from the new JSON format...
}

// Message represents the structure of a message within a session.
type Message struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Role    string `json:"role"`
	Content string `json:"content"`
	// Include other fields from the message structure...
}

// Stat represents the structure of the stat field within a session.
type Stat struct {
	TokenCount int `json:"tokenCount"`
	WordCount  int `json:"wordCount"`
	CharCount  int `json:"charCount"`
	// Include other fields from the stat structure...
}

// Mask represents the structure of the mask field within a session.
type Mask struct {
	ID               StringOrInt  `json:"id"`
	Avatar           string       `json:"avatar"`
	Name             string       `json:"name"`
	Context          []Message    `json:"context"`
	SyncGlobalConfig bool         `json:"syncGlobalConfig"`
	ModelConfig      *ModelConfig `json:"modelConfig"`
	Lang             string       `json:"lang"`
	Builtin          bool         `json:"builtin"`
	CreatedAt        int64        `json:"createdAt"`
	// Include other fields from the mask structure...
}

// ModelConfig represents the structure of the modelConfig field within a mask.
type ModelConfig struct {
	Model                          string        `json:"model"`                          // The model name.
	Temperature                    float64       `json:"temperature"`                    // The temperature for generating responses.
	TopP                           float64       `json:"top_p"`                          // The top-p value for generating responses.
	MaxTokens                      int           `json:"max_tokens"`                     // The maximum number of tokens in a generated response.
	PresencePenalty                float64       `json:"presence_penalty"`               // The presence penalty for generating responses.
	FrequencyPenalty               float64       `json:"frequency_penalty"`              // The frequency penalty for generating responses.
	N                              int           `json:"n"`                              // The number of responses to generate.
	Quality                        string        `json:"quality"`                        // The quality of the generated responses.
	Size                           string        `json:"size"`                           // The size of the model.
	Style                          string        `json:"style"`                          // The style of the generated responses.
	SystemFingerprint              string        `json:"system_fingerprint"`             // The fingerprint of the system.
	SendMemory                     bool          `json:"sendMemory"`                     // Whether to send memory to the model.
	HistoryMessageCount            int           `json:"historyMessageCount"`            // The number of history messages to include.
	CompressMessageLengthThreshold int           `json:"compressMessageLengthThreshold"` // The threshold for compressing message length.
	EnableInjectSystemPrompts      bool          `json:"enableInjectSystemPrompts"`      // Whether to enable injecting system prompts.
	Template                       string        `json:"template"`                       // The template for generating responses.
	SystemPrompt                   *SystemPrompt `json:"systemprompt,omitempty"`         // The system prompt for generating responses (optional).
}

// SystemPrompt represents the structure of the systemprompt field within a modelConfig.
type SystemPrompt struct {
	Default string `json:"default"`
}

// RepairSessionData takes a byte slice of the old JSON format and returns a byte slice of the new JSON format.
type Session struct {
	ID                 string    `json:"id"`
	Topic              string    `json:"topic"`
	MemoryPrompt       string    `json:"memoryPrompt"`
	Messages           []Message `json:"messages"`
	Stat               Stat      `json:"stat"`
	LastUpdate         int64     `json:"lastUpdate"`
	LastSummarizeIndex int       `json:"lastSummarizeIndex"`
	Mask               *Mask     `json:"mask"`
}

// RepairSessionData transforms JSON data from the old format to the new format.
//
// It adds a 'systemprompt' field to the 'modelConfig' within each session if it is missing.
func RepairSessionData(oldDataBytes []byte) ([]byte, error) {
	var oldData OldData
	err := json.Unmarshal(oldDataBytes, &oldData)
	if err != nil {
		return nil, err
	}

	// Initialize the new data structure with the old data.
	newData := NewData{
		ChatNextWebStore: oldData.ChatNextWebStore,
	}
	// Iterate through the sessions to copy and transform each one.
	for i, session := range newData.ChatNextWebStore.Sessions {
		// Check if the systemprompt field is missing and add it if necessary.
		if session.Mask != nil && session.Mask.ModelConfig != nil && session.Mask.ModelConfig.SystemPrompt == nil {
			newData.ChatNextWebStore.Sessions[i].Mask.ModelConfig.SystemPrompt = &SystemPrompt{
				Default: "\nYou are ChatGPT, a large language model trained by OpenAI.\nKnowledge cutoff: {{cutoff}}\nCurrent model: {{model}}\nCurrent time: {{time}}\nLatex inline: $x^2$ \nLatex block: $$e=mc^2$$\n",
			}
		}
	}

	// Marshal the new data into JSON bytes.
	newDataBytes, err := json.MarshalIndent(newData, "", "  ")
	if err != nil {
		return nil, err
	}

	return newDataBytes, nil
}

// Helper function millisToTime converts Unix milliseconds to a time.Time object.
// This is used to handle date and time fields in the JSON data that are represented as Unix millisecond timestamps.
func millisToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
