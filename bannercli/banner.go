// banner/banner.go
// Copyright (c) 2023 H0llyW00dzZ
package banner

import (
	"fmt"
	"strings"
	"time"
)

// PrintBinaryBanner prints a binary representation of a banner.
func PrintBinaryBanner(message string) {
	banner := strings.ReplaceAll(message, " ", "   ")
	for _, char := range banner {
		fmt.Printf(" %08b", char)
	}
	fmt.Println()
}

// PrintAnimatedBanner prints a simple animated banner by scrolling the message.
func PrintAnimatedBanner(message string, repeat int, delay time.Duration) {
	for r := 0; r < repeat; r++ {
		for i := 0; i < len(message); i++ {
			fmt.Print("\r" + strings.Repeat(" ", i) + message)
			time.Sleep(delay)
		}
	}
	fmt.Println()
}
