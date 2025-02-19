package transfer

import (
    "bytes"
    "fmt"
    "io"
    "io/ioutil"
    "mime/multipart"
    "net/http"
    "os"
    "path"
    "client/utils"
)

// UploadWebhook uploads files to a webhook via HTTP POST request
func UploadWebhook(localPath, webhookURL string, includePatterns []string, scheduleInterval int) error {
    return utils.WalkAndUpload(localPath, "", includePatterns, func(localFilePath, _ string) error {
        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to stat file: %s", err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
            return nil
        }

        fmt.Printf("[*] Attempting to upload file: %s to webhook: %s\n", localFilePath, webhookURL)

        err = uploadFileWebhook(localFilePath, webhookURL, fileInfo)
        if err != nil {
            return fmt.Errorf("\n[-] Error uploading file: %s", err)
        }

        return nil
    }, scheduleInterval)
}

// uploadFileWebhook handles the actual file upload to the webhook
func uploadFileWebhook(localFilePath, webhookURL string, fileInfo os.FileInfo) error {
    file, err := os.Open(localFilePath)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to open file: %s", err)
    }
    defer file.Close()

    var requestBody bytes.Buffer
    writer := multipart.NewWriter(&requestBody)
    part, err := writer.CreateFormFile("file", path.Base(localFilePath))
    if err != nil {
        return fmt.Errorf("\n[-] Failed to create form file: %s", err)
    }

    // Progress tracking
    buf := make([]byte, 1024*1024)
    var total int64
    totalSize := fileInfo.Size()

    for {
        n, err := file.Read(buf)
        if err != nil && err != io.EOF {
            return err
        }
        if n == 0 {
            break
        }

        if _, err := part.Write(buf[:n]); err != nil {
            return err
        }

        total += int64(n)
        progress := float64(total) / float64(totalSize) * 100
        fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, progress)
    }

    err = writer.Close()
    if err != nil {
        return fmt.Errorf("\n[-] Failed to close writer: %s", err)
    }

    req, err := http.NewRequest("POST", webhookURL, &requestBody)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to create request: %s", err)
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to send request: %s", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return fmt.Errorf("\n[-] Request failed with status: %s, body: %s", resp.Status, string(body))
    }

    return nil
}
