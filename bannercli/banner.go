// Copyright (c) 2023 H0llyW00dzZ
//
// Package bannercli provides functionality to print different styles of banners
// to the terminal. These styles include binary representation and simple
// animation effects to enhance the visual presentation of CLI applications.
//
// # Example Usage
//
//	func main() {
//		// ... existing code ...
//
//		// Display a binary banner
//		banner.PrintTypingBanner("ChatGPT Session Exporter", 100*time.Millisecond)
//		// Optionally, display an typing animated banner
//		banner.PrintAnimatedBanner("ChatGPT Session Exporter", 3, 200*time.Millisecond)
//
//		// ... rest of your main function ...
//	}
package bannercli

import (
	"fmt"
	"strings"
	"time"
)

// PrintBinaryBanner prints a binary representation of a banner.
// Each character of the message is converted into its binary form.
// Spaces between words are widened to enhance readability.
func PrintBinaryBanner(message string) {
	banner := strings.ReplaceAll(message, " ", "   ")
	for _, char := range banner {
		fmt.Printf(" %08b", char)
	}
	fmt.Println()
}

// PrintAnimatedBanner prints a simple animated banner by scrolling the message
// horizontally across the terminal. The animation repeats the number of times
// specified by the `repeat` parameter with a delay between each frame as
// specified by the `delay` parameter.
func PrintAnimatedBanner(message string, repeat int, delay time.Duration) {
	for r := 0; r < repeat; r++ {
		for i := 0; i < len(message); i++ {
			fmt.Print("\r" + strings.Repeat(" ", i) + message)
			time.Sleep(delay)
		}
	}
	fmt.Println()
}

// PrintTypingBanner prints the message with a typing animation effect.
//
// Each character appears sequentially with a delay, simulating a typing effect.
//
// Note: This simulation typing just like a human would type.
func PrintTypingBanner(message string, delay time.Duration) {
	for _, char := range message {
		fmt.Printf("%c", char)
		time.Sleep(delay)
	}
	fmt.Println()
}
