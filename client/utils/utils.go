package utils

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "path"
    "time"
)

// WalkAndUpload processes files recursively with optional scheduling
func WalkAndUpload(localPath, remoteDir string, includePatterns []string, uploadFunc func(string, string) error, scheduleInterval int) error {
    var filesToUpload []struct {
        LocalPath  string
        RemotePath string
    }

    err := filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Get the relative path for remote directory structure
        relativePath, relErr := filepath.Rel(localPath, filePath)
        if relErr != nil {
            return relErr
        }
        remoteFilePath := path.Join(remoteDir, relativePath) // Use relativePath

        if info.IsDir() {
            fmt.Printf("[*] Processing directory: %s\n", filePath)
            return nil // Skip directory uploads
        }

        // Apply include pattern filtering
        if hasValidPatterns(includePatterns) && !matchesIncludePatterns(filePath, includePatterns) {
            fmt.Printf("[!] Skipping file: %s (does not match include patterns)\n", filePath)
            return nil
        }

        // Store file with its respective remote path
        filesToUpload = append(filesToUpload, struct {
            LocalPath  string
            RemotePath string
        }{filePath, remoteFilePath})

        return nil
    })

    if err != nil {
        return err
    }

    if len(filesToUpload) == 0 {
        fmt.Println("[-] No matching files found for upload.")
        return nil
    }

    // If scheduling is enabled, upload files one by one at intervals
    if scheduleInterval > 0 {
        fmt.Printf("[*] Uploading each file every %d minutes...\n", scheduleInterval)

        for _, file := range filesToUpload {
            fmt.Printf("[*] Uploading file: %s to %s\n", file.LocalPath, file.RemotePath)

            err := uploadFunc(file.LocalPath, file.RemotePath)
            if err != nil {
                fmt.Printf("[-] Error uploading %s: %s\n", file.LocalPath, err)
            }  else {
                fmt.Printf("[+] Successfully uploaded: %s\n", file.LocalPath)
            }

            fmt.Printf("[*] Waiting %d minutes before uploading next file...\n", scheduleInterval)

            time.Sleep(time.Duration(scheduleInterval) * time.Minute) // Wait before next upload
        }
    } else {
        // Immediate upload if no schedule is set
        for _, file := range filesToUpload {
            fmt.Printf("[*] Uploading file: %s to %s\n", file.LocalPath, file.RemotePath)

            err := uploadFunc(file.LocalPath, file.RemotePath)
            if err != nil {
                fmt.Printf("[-] Error uploading %s: %s\n", file.LocalPath, err)
            } else {
                fmt.Printf("[+] Successfully uploaded: %s\n", file.LocalPath)
            }
        }
    }

    return nil
}

// Function to check if a file matches any of the include patterns
func matchesIncludePatterns(filePath string, includePatterns []string) bool {
    fileName := filepath.Base(filePath)

    for _, pattern := range includePatterns {
        pattern = strings.TrimSpace(pattern) // Handle spaces in patterns
        if pattern == "" { // Skip empty patterns
            continue
        }

        matched, err := filepath.Match(pattern, fileName)
        if err != nil {
            fmt.Printf("[!] Error matching pattern '%s': %v\n", pattern, err)
            continue
        }

        if matched {
            return true
        }
    }
    return false
}

// Helper function to check if includePatterns has any valid (non-empty) pattern
func hasValidPatterns(patterns []string) bool {
    for _, pattern := range patterns {
        if strings.TrimSpace(pattern) != "" {
            return true
        }
    }
    return false
}
