// Copyright (c) 2023 H0llyW00dzZ
//
// Package updater provides functionality to automatically update a Go application
// by checking for the latest release on GitHub and, if available, downloading and
// applying the update. It is designed to work with applications that are distributed
// with GitHub releases.
//
// The updater checks the latest release by calling the GitHub Releases API and
// compares the tag name of the latest release with the current version of the
// application. If the tag name indicates a newer version, the updater downloads
// the release asset that matches the running application's operating system and
// architecture, replaces the current executable, and restarts the application.
//
// Usage:
//
// To use the updater, you should include it in your application's main package:
//
//	import "github.com/H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter/updater"
//
//	func main() {
//	    if err := updater.UpdateApplication(); err != nil {
//	        // Handle error
//	    }
//	    // Continue with application logic
//	}
//
// The updater assumes that the GitHub repository's release assets follow a
// naming convention that includes the OS and architecture. It also assumes that
// the binary to be updated is named "myapp" and is located in the current working
// directory of the running application.
//
// Note that the updater package defines a constant `currentVersion` that must
// be updated to match the application's current version string before building
// a new release. This version string is used to compare against the tag name of
// the latest release on GitHub.
//
// The updater package is designed with simplicity in mind and does not handle
// complex update scenarios such as database migrations, configuration changes,
// or rollback of failed updates. It is recommended to test the update process
// thoroughly in a controlled environment before deploying it in a production setting.
//
// Security Considerations:
//
// The updater performs a direct binary replacement and restarts the application.
// Users should ensure that the GitHub repository and release assets are secure
// and that the release process includes steps to verify the integrity and
// authenticity of the binaries, such as signing the releases.
//
// # Additional Note: This Package Currently under development.
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const (
	currentVersion = "1.3.3.7"
	githubRepo     = "H0llyW00dzZ/ChatGPT-Next-Web-Session-Exporter"
)

// releaseInfo defines the structure for storing information about a GitHub release.
// It captures the tag name of the release and a slice of assets that are part of the release.
type releaseInfo struct {
	TagName string `json:"tag_name"` // The name of the tag for the release.
	Assets  []struct {
		Name               string `json:"name"`                 // The name of the asset.
		BrowserDownloadURL string `json:"browser_download_url"` // The URL for downloading the asset.
	} `json:"assets"` // A list of assets available for the release.
}

// getLatestRelease fetches the latest release information from the GitHub repository.
// It constructs a request to the GitHub API to retrieve the latest release and parses
// the response into a releaseInfo struct.
//
// Returns a pointer to a releaseInfo struct and nil error on success.
// On failure, it returns nil and an error indicating what went wrong.
func getLatestRelease() (*releaseInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API response status: %s", resp.Status)
	}

	var release releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// UpdateApplication checks the GitHub repository for a newer release of the application.
// If a newer release is found, it downloads the corresponding binary for the current
// platform and architecture, replaces the current executable with the downloaded binary,
// and restarts the application.
//
// Returns nil if the application is up to date or the update is successfully applied.
// If an error occurs during the update process, it returns a non-nil error.
func UpdateApplication() error {
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("error fetching latest release: %w", err)
	}

	if release.TagName == currentVersion {
		fmt.Println("No update available.")
		return nil
	}

	fmt.Printf("Update available: %s\n", release.TagName)
	fmt.Println("Downloading update...")

	// Find the asset that matches our platform
	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == fmt.Sprintf("ChatGPT-Next-Web-Session-Exporter-%s-%s", runtime.GOOS, runtime.GOARCH) {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no binary for the current platform")
	}

	// Download the new binary
	resp, err := http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("error downloading update: %w", err)
	}
	defer resp.Body.Close()

	// Create a new file with a temporary name
	out, err := os.CreateTemp("", "ChatGPT-Next-Web-Session-Exporter-*")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer out.Close()

	// Write the downloaded content to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing update to file: %w", err)
	}

	// Close the file before renaming it
	if err := out.Close(); err != nil {
		return fmt.Errorf("error closing update file: %w", err)
	}

	fmt.Println("Update downloaded. Applying update...")

	// Replace the current binary with the new one
	if err := os.Rename(out.Name(), "ChatGPT-Next-Web-Session-Exporter"); err != nil {
		return fmt.Errorf("error replacing binary: %w", err)
	}

	fmt.Println("Update applied. Restarting application...")

	// Restart the application
	cmd := exec.Command("ChatGPT-Next-Web-Session-Exporter")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error restarting application: %w", err)
	}

	// Exit the current process
	os.Exit(0)

	return nil
}
