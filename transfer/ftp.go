package transfer

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "strings"
    "VeilTransfer/utils"
    "github.com/jlaffaye/ftp"
)

func UploadFTP(username, password, server, localPath, remoteDir string, includePatterns []string) error {
    conn, err := ftp.Dial(server)
    if err != nil {
        return fmt.Errorf("[-] failed to connect: %s", err)
    }

    err = conn.Login(username, password)
    if err != nil {
        return fmt.Errorf("[-] failed to login: %s", err)
    }
    defer conn.Logout()

    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/")

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
            return nil
        }

        srcFile, err := os.Open(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to open source file: %s", err)
        }
        defer srcFile.Close()

        buf := make([]byte, 1024*1024)
        var total int64
        totalSize := fileInfo.Size()

        progress := make(chan int, 1)

        go func() {
            for p := range progress {
                percentage := float64(p) / float64(totalSize) * 100
                fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
            }
        }()

        for {
            n, err := srcFile.Read(buf)
            if err != nil && err != io.EOF {
                return err
            }
            if n == 0 {
                break
            }

            if err := conn.Stor(remoteFilePath, bytes.NewReader(buf[:n])); err != nil {
                return fmt.Errorf("[-] failed to upload file: %s", err)
            }

            total += int64(n)
            progress <- int(total)
        }

        close(progress)
        return nil
    })
}