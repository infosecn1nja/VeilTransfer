package transfer

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func UploadGithub(token, localPath string, repo string, includePatterns []string) error {
    // Prepare the URL and repository path
    baseURL := "https://api.github.com/repos/" + repo + "/contents/"

    return WalkAndUpload(localPath, includePatterns, func(localFilePath, remoteFilePath string) error {
        remoteFilePath = filepath.Base(remoteFilePath)

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
            return nil
        }

        // Read the file
        fileData, err := ioutil.ReadFile(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to read file: %s", err)
        }

        // Encode the file content to base64
        content := base64.StdEncoding.EncodeToString(fileData)

        // Create the JSON payload
        payload := map[string]string{
            "message": "Upload via VeilTransfer",
            "content": content,
        }
        payloadBytes, err := json.Marshal(payload)
        if err != nil {
            return fmt.Errorf("[-] failed to marshal JSON payload: %s", err)
        }

        // Create the full URL for the file to be uploaded
        uploadURL := baseURL + remoteFilePath

        // Create the HTTP request
        req, err := http.NewRequest("PUT", uploadURL, strings.NewReader(string(payloadBytes)))
        if err != nil {
            return fmt.Errorf("[-] failed to create HTTP request: %s", err)
        }

        // Set the headers
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")

        // Make the request
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            return fmt.Errorf("[-] failed to upload file: %s", err)
        }
        defer resp.Body.Close()

        // Check for a successful response
        if resp.StatusCode != http.StatusCreated {
            body, _ := ioutil.ReadAll(resp.Body)
            return fmt.Errorf("[-] failed to upload file: %s, response: %s", resp.Status, string(body))
        }

        fmt.Printf("[*] Successfully uploaded %s to GitHub repository\n", localFilePath)

        return nil
    })
}

func WalkAndUpload(localPath string, includePatterns []string, uploadFunc func(localFilePath, remoteFilePath string) error) error {
    var errors []error

    filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            fmt.Printf("[-] Error walking path: %s\n", err)
            return nil // Continue even if there's an error
        }

        // If includePatterns is not provided or empty, upload all files
        if len(includePatterns) > 0 {
            matched := false
            for _, pattern := range includePatterns {
                matched, err = filepath.Match(pattern, filepath.Base(filePath))
                if err != nil {
                    fmt.Printf("[-] Error matching pattern: %s\n", err)
                    return nil // Continue even if there's an error
                }
                if matched {
                    break
                }
            }
            if !matched {
                return nil // Skip this file if it doesn't match any pattern
            }
        }

        remoteFilePath := strings.TrimPrefix(filePath, localPath+"/")
        if err := uploadFunc(filePath, remoteFilePath); err != nil {
            fmt.Printf("[-] Error uploading file %s: %s\n", filePath, err)
            errors = append(errors, err) // Collect the error and continue
        }

        return nil
    })

    if len(errors) > 0 {
        return fmt.Errorf("encountered %d errors during upload", len(errors))
    }

    return nil
}