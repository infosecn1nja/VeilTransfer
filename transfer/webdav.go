package transfer

import (
    "bytes"
    "fmt"   
    "os"
    "io"
    "path/filepath"
    "github.com/studio-b12/gowebdav"
)

func UploadWebDAV(username, password, server, localPath, remoteDir string, includePatterns []string) error {
    client := gowebdav.NewClient(server, username, password)

    err := client.Connect()
    if err != nil {
        return fmt.Errorf("[-] failed to connect: %s", err)
    }

    err = filepath.Walk(localPath, func(localFilePath string, info os.FileInfo, err error) error {
        if err != nil {
            return fmt.Errorf("[-] error accessing path %q: %s", localFilePath, err)
        }

        if info.IsDir() {
            return nil
        }

        relPath, err := filepath.Rel(localPath, localFilePath)
        if err != nil {
            return fmt.Errorf("[-] error calculating relative path: %s", err)
        }

        remoteFilePath := filepath.ToSlash(filepath.Join(remoteDir, relPath))

        remoteDirPath := filepath.ToSlash(filepath.Dir(remoteFilePath))
        err = client.MkdirAll(remoteDirPath, 0755)
        if err != nil {
            return fmt.Errorf("[-] failed to create remote directory %q: %s", remoteDirPath, err)
        }

        srcFile, err := os.Open(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to open source file: %s", err)
        }
        defer srcFile.Close()

        buf := make([]byte, 1024*1024)
        var total int64
        totalSize := info.Size()

        for {
            n, err := srcFile.Read(buf)
            if err != nil && err != io.EOF {
                return fmt.Errorf("[-] error reading file: %s", err)
            }
            if n == 0 {
                break
            }

            if err := client.WriteStream(remoteFilePath, bytes.NewReader(buf[:n]), 0664); err != nil {
                return fmt.Errorf("[-] failed to upload file: %s", err)
            }

            total += int64(n)
            progress := float64(total) / float64(totalSize) * 100
            fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, progress)
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("[-] Error during WebDAV upload: %s", err)
    }

    return nil
}