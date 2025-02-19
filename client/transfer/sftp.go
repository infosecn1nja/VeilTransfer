package transfer

import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "strings"
    "client/utils"
    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

// UploadSFTP uploads files to an SFTP server
func UploadSFTP(username, password, server, localPath, remoteDir, privateKeyPath string, includePatterns []string, scheduleInterval int) error {
    var authMethods []ssh.AuthMethod

    if privateKeyPath != "" {
        key, err := ioutil.ReadFile(privateKeyPath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to read private key file: %s", err)
        }

        // Attempt to parse the private key with or without a passphrase
        signer, err := ssh.ParsePrivateKey(key)
        if err != nil {
            if _, ok := err.(*ssh.PassphraseMissingError); ok {
                return fmt.Errorf("\n[-] Private key is encrypted and no passphrase was provided")
            }
            return fmt.Errorf("\n[-] Failed to parse private key: %s", err)
        }

        authMethods = append(authMethods, ssh.PublicKeys(signer))
    } else {
        authMethods = append(authMethods, ssh.Password(password))
    }

    sshConfig := &ssh.ClientConfig{
        User:            username,
        Auth:            authMethods,
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    conn, err := ssh.Dial("tcp", server, sshConfig)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to dial: %s", err)
    }
    defer conn.Close()

    client, err := sftp.NewClient(conn)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to create SFTP client: %s", err)
    }
    defer client.Close()

    if err := ensureRemoteDir(client, remoteDir); err != nil {
        return fmt.Errorf("\n[-] Failed to ensure remote directory: %s", err)
    }

    // Use WalkAndUpload for recursive directory traversal and scheduling
    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        localFilePath = strings.ReplaceAll(localFilePath, "\\", "/")
        remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/")

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            remoteDirPath := remoteFilePath
            if err := client.MkdirAll(remoteDirPath); err != nil {
                return fmt.Errorf("\n[-] Failed to create remote directory: %s", err)
            }
            return nil
        }

        fmt.Printf("[*] Uploading %s to %s\n", localFilePath, remoteFilePath)

        err = uploadFileSFTP(client, localFilePath, remoteFilePath, fileInfo)
        if err != nil {
            return fmt.Errorf("\n[-] Error uploading file: %s", err)
        }

        return nil
    }, scheduleInterval)
}

// uploadFileSFTP uploads a single file to SFTP with progress tracking
func uploadFileSFTP(client *sftp.Client, localFilePath, remoteFilePath string, fileInfo os.FileInfo) error {
    srcFile, err := os.Open(localFilePath)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to open source file: %s", err)
    }
    defer srcFile.Close()

    dstFile, err := client.Create(remoteFilePath)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to create destination file: %s", err)
    }
    defer dstFile.Close()

    buf := make([]byte, 1024*1024)
    var total int64
    totalSize := fileInfo.Size()

    for {
        n, err := srcFile.Read(buf)
        if err != nil && err != io.EOF {
            return err
        }
        if n == 0 {
            break
        }

        if _, err := dstFile.Write(buf[:n]); err != nil {
            return fmt.Errorf("\n[-] Failed to write to destination file: %s", err)
        }

        total += int64(n)
        percentage := float64(total) / float64(totalSize) * 100
        fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
    }

    return nil
}

// ensureRemoteDir ensures the remote directory exists
func ensureRemoteDir(client *sftp.Client, remoteDir string) error {
    remoteDir = strings.ReplaceAll(remoteDir, "\\", "/")

    if _, err := client.Stat(remoteDir); err != nil {
        if os.IsNotExist(err) {
            if err := client.MkdirAll(remoteDir); err != nil {
                return fmt.Errorf("\n[-] Failed to create remote directory: %s", err)
            }
        } else {
            return err
        }
    }
    return nil
}
