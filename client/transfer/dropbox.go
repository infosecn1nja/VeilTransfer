package transfer

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "client/utils"
)

// DropboxAPIArgs defines the required structure for Dropbox API upload requests
type DropboxAPIArgs struct {
    Path           string `json:"path"`
    Mode           string `json:"mode"`
    Autorename     bool   `json:"autorename"`
    Mute           bool   `json:"mute"`
    StrictConflict bool   `json:"strict_conflict"`
}

// UploadDropbox uploads files to Dropbox with optional scheduling
func UploadDropbox(accessToken, localPath, remoteDir string, includePatterns []string, scheduleInterval int) error {
    // Define the upload function for Dropbox
    uploadFunc := func(localFilePath, remoteFilePath string) error {
        if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
            return fmt.Errorf("File does not exist: %s", localFilePath)
        }

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to stat file %s: %v", localFilePath, err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
            return nil
        }

        fileData, err := ioutil.ReadFile(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to read file %s: %v", localFilePath, err)
        }

        // Construct Dropbox upload request body
        dropboxArgs := DropboxAPIArgs{
            Path:           remoteFilePath,
            Mode:           "add",
            Autorename:     true,
            Mute:           false,
            StrictConflict: false,
        }

        dropboxArgsJSON, _ := json.Marshal(dropboxArgs)

        req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", bytes.NewBuffer(fileData))
        if err != nil {
            return fmt.Errorf("\n[-] Failed to create request for %s: %v", localFilePath, err)
        }

        // Set required headers for Dropbox API
        req.Header.Set("Authorization", "Bearer "+accessToken)
        req.Header.Set("Content-Type", "application/octet-stream")
        req.Header.Set("Dropbox-API-Arg", string(dropboxArgsJSON))

        // Execute the request
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to upload file %s: %v", localFilePath, err)
        }
        defer resp.Body.Close()

        // Handle unsuccessful responses
        if resp.StatusCode != http.StatusOK {
            body, _ := ioutil.ReadAll(resp.Body)
            return fmt.Errorf("\n[-] Upload failed for %s with status: %s, response: %s", localFilePath, resp.Status, string(body))
        }

        return nil
    }

    // Use WalkAndUpload for recursive directory traversal & scheduling
    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        relPath, err := filepath.Rel(localPath, localFilePath)
        if err != nil {
            return fmt.Errorf("Failed to get relative path: %v", err)
        }
        remoteFilePath = filepath.Join(remoteDir, relPath)
        remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/") // Ensure correct Dropbox path format

        return uploadFunc(localFilePath, remoteFilePath)
    }, scheduleInterval)
}
