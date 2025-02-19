package transfer

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "client/utils"
    "github.com/jlaffaye/ftp"
)

// UploadFTP uploads files to an FTP server with optional scheduling
func UploadFTP(username, password, server, localPath, remoteDir string, includePatterns []string, scheduleInterval int) error {
    conn, err := ftp.Dial(server)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to connect to FTP server: %s", err)
    }

    err = conn.Login(username, password)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to login to FTP server: %s", err)
    }
    defer conn.Logout()

    // Use WalkAndUpload for recursive directory traversal & scheduling
    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/")

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[*] Creating remote directory: %s\n", remoteFilePath)
            if err := conn.MakeDir(remoteFilePath); err != nil && !strings.Contains(err.Error(), "550") {
                return fmt.Errorf("\n[-] Failed to create remote directory: %s", err)
            }
            return nil
        }

        srcFile, err := os.Open(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to open source file: %s", err)
        }
        defer srcFile.Close()

        totalSize := fileInfo.Size()

        progress := make(chan int, 1)
        go func() {
            for p := range progress {
                percentage := float64(p) / float64(totalSize) * 100
                fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
            }
        }()

        remoteFilePath = filepath.Join(remoteDir, filepath.Base(localFilePath))
        remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/")

        if err := conn.Stor(remoteFilePath, srcFile); err != nil {
            return fmt.Errorf("\n[-] Failed to upload file: %s", err)
        }

        progress <- int(totalSize)
        close(progress)

        return nil
    }, scheduleInterval)
}
