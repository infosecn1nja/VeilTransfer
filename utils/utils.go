package utils

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "path"
)

// Recursive directory walk and upload with file extension filtering
func WalkAndUpload(localPath, remoteDir string, includePatterns []string, uploadFunc func(string, string) error) error {
    return filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        relativePath := strings.TrimPrefix(filePath, localPath)
        remoteFilePath := path.Join(remoteDir, relativePath)

        if info.IsDir() {
            fmt.Printf("[*] Processing directory: %s\n", filePath)
            if err := uploadFunc(filePath, remoteFilePath); err != nil {
                return err
            }
            return nil
        }

        if !matchesIncludePatterns(filePath, includePatterns) {
            fmt.Printf("[!] Skipping file: %s (does not match include patterns)\n", filePath)
            return nil
        }

        fmt.Printf("[*] Processing file: %s\n", filePath)
        return uploadFunc(filePath, remoteFilePath)
    })
}

// Function to check if a file matches any of the include patterns
func matchesIncludePatterns(filePath string, includePatterns []string) bool {
    if len(includePatterns) == 0 {
        return true
    }
    for _, pattern := range includePatterns {
        matched, _ := filepath.Match(pattern, filepath.Base(filePath))
        if matched {
            return true
        }
    }
    return false
}
