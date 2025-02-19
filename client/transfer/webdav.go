package transfer

import (
    "fmt"
    "os"
    "io"
    "path"
    "client/utils"
    "github.com/studio-b12/gowebdav"
)

// UploadWebDAV uploads files to a WebDAV server
func UploadWebDAV(username, password, server, localPath, remoteDir string, includePatterns []string, scheduleInterval int) error {
    client := gowebdav.NewClient(server, username, password)

    err := client.Connect()
    if err != nil {
        return fmt.Errorf("\n[-] Failed to connect: %s", err)
    }

    processedFiles := make(map[string]bool) // Track uploaded files

    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        if processedFiles[localFilePath] {
            return nil // Skip duplicate uploads
        }

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Error accessing file %q: %s", localFilePath, err)
        }

        // Ensure correct remote path
        remoteFilePath = path.Join(remoteDir, path.Base(localFilePath))


        err = uploadFile(client, localFilePath, remoteFilePath, fileInfo)
        if err != nil {
            return fmt.Errorf("\n[-] Error uploading file %q: %s", localFilePath, err)
        }

        processedFiles[localFilePath] = true // Mark as uploaded
        return nil
    }, scheduleInterval)
}

// uploadFile uploads a file to WebDAV using PUT
func uploadFile(client *gowebdav.Client, localFilePath, remoteFilePath string, fileInfo os.FileInfo) error {
    srcFile, err := os.Open(localFilePath)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to open source file %q: %s", localFilePath, err)
    }
    defer srcFile.Close()

    data, err := io.ReadAll(srcFile)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to read file %q: %s", localFilePath, err)
    }

    // Upload file
    err = client.Write(remoteFilePath, data, 0644)
    if err != nil {
        return fmt.Errorf("\n[-] PUT failed for file %q: %s", localFilePath, err)
    }

    return nil
}
