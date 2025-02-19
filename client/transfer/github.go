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
    "client/utils"
)

// UploadGithub uploads files to a GitHub repository with optional scheduling
func UploadGithub(token, localPath string, repo string, includePatterns []string, scheduleInterval int) error {
    // Prepare the base API URL
    baseURL := "https://api.github.com/repos/" + repo + "/contents/"

    // Use WalkAndUpload for recursive directory traversal & scheduling
    return utils.WalkAndUpload(localPath, "", includePatterns, func(localFilePath, remoteFilePath string) error {
        remoteFilePath = filepath.ToSlash(remoteFilePath) // Ensure consistent path format for GitHub

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
            return nil
        }

        // Read the file
        fileData, err := ioutil.ReadFile(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to read file: %s", err)
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
            return fmt.Errorf("\n[-] Failed to marshal JSON payload: %s", err)
        }

        // Create the full URL for the file to be uploaded
        uploadURL := baseURL + remoteFilePath

        // Create the HTTP request
        req, err := http.NewRequest("PUT", uploadURL, strings.NewReader(string(payloadBytes)))
        if err != nil {
            return fmt.Errorf("\n[-] Failed to create HTTP request: %s", err)
        }

        // Set the headers
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")

        // Make the request
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to upload file: %s", err)
        }
        defer resp.Body.Close()

        // Check for a successful response
        if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
            body, _ := ioutil.ReadAll(resp.Body)
            return fmt.Errorf("\n[-] Failed to upload file: %s, response: %s", resp.Status, string(body))
        }

        return nil
    }, scheduleInterval)
}
