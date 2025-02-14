package transfer

import (
    "bytes"
    "fmt"
    "os"
    "io"
    "path/filepath"
    "time"
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

    // Use WalkAndUpload for recursive directory traversal & scheduling
    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Error accessing file %q: %s", localFilePath, err)
        }

        fmt.Printf("[*] Uploading file: %s to %s\n", localFilePath, remoteFilePath)

        err = uploadFile(client, localFilePath, remoteFilePath, fileInfo)
        if err != nil {
            return fmt.Errorf("\n[-] Error uploading file: %s", err)
        }

        fmt.Printf("[+] Successfully uploaded: %s\n", localFilePath)

        // If scheduling is enabled, wait before the next upload
        if scheduleInterval > 0 {
            fmt.Printf("[*] Waiting %d minutes before uploading next file...\n", scheduleInterval)
            time.Sleep(time.Duration(scheduleInterval) * time.Minute)
        }

        return nil
    }, scheduleInterval)
}

// uploadFile uploads a single file to WebDAV with progress tracking
func uploadFile(client *gowebdav.Client, localFilePath, remoteFilePath string, fileInfo os.FileInfo) error {
    remoteDirPath := filepath.ToSlash(filepath.Dir(remoteFilePath))

    // Create remote directory if it doesn't exist
    err := client.MkdirAll(remoteDirPath, 0755)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to create remote directory %q: %s", remoteDirPath, err)
    }

    srcFile, err := os.Open(localFilePath)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to open source file: %s", err)
    }
    defer srcFile.Close()

    buf := make([]byte, 1024*1024)
    var total int64
    totalSize := fileInfo.Size()

    for {
        n, err := srcFile.Read(buf)
        if err != nil && err != io.EOF {
            return fmt.Errorf("\n[-] Error reading file: %s", err)
        }
        if n == 0 {
            break
        }

        if err := client.WriteStream(remoteFilePath, bytes.NewReader(buf[:n]), 0664); err != nil {
            return fmt.Errorf("\n[-] Failed to upload file: %s", err)
        }

        total += int64(n)
        progress := float64(total) / float64(totalSize) * 100
        fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, progress)
    }

    return nil
}
