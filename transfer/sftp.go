package transfer

import (
    "fmt"
    "io"
    "io/ioutil"
    "strings"
    "os"
    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
    "VeilTransfer/utils"
)

func UploadSFTP(username, password, server, localPath, remoteDir, privateKeyPath string, includePatterns []string) error {
    var authMethods []ssh.AuthMethod

    if privateKeyPath != "" {
        key, err := ioutil.ReadFile(privateKeyPath)
        if err != nil {
            return fmt.Errorf("[-] failed to read private key file: %s", err)
        }

        // Attempt to parse the private key with or without a passphrase
        var signer ssh.Signer
        signer, err = ssh.ParsePrivateKey(key)
        if err != nil {
            if _, ok := err.(*ssh.PassphraseMissingError); ok {
                return fmt.Errorf("[-] private key is encrypted and no passphrase was provided")
            }
            return fmt.Errorf("[-] failed to parse private key: %s", err)
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
        return fmt.Errorf("[-] failed to dial: %s", err)
    }
    defer conn.Close()

    client, err := sftp.NewClient(conn)
    if err != nil {
        return fmt.Errorf("[-] failed to create sftp client: %s", err)
    }
    defer client.Close()

    if err := ensureRemoteDir(client, remoteDir); err != nil {
        return fmt.Errorf("[-] failed to ensure remote directory: %s", err)
    }

    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        localFilePath = strings.ReplaceAll(localFilePath, "\\", "/")
        remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/")

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            remoteDirPath := remoteFilePath
            if err := client.MkdirAll(remoteDirPath); err != nil {
                return fmt.Errorf("[-] failed to create remote directory: %s", err)
            }
            return nil
        }

        fmt.Printf("[*] Uploading %s to %s\n", localFilePath, remoteFilePath)

        srcFile, err := os.Open(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to open source file: %s", err)
        }
        defer srcFile.Close()

        dstFile, err := client.Create(remoteFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to create destination file: %s", err)
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
                return fmt.Errorf("[-] failed to write to destination file: %s", err)
            }

            total += int64(n)
            percentage := float64(total) / float64(totalSize) * 100
            fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
        }

        return nil
    })
}

func ensureRemoteDir(client *sftp.Client, remoteDir string) error {
    remoteDir = strings.ReplaceAll(remoteDir, "\\", "/")

    if _, err := client.Stat(remoteDir); err != nil {
        if os.IsNotExist(err) {
            if err := client.MkdirAll(remoteDir); err != nil {
                return fmt.Errorf("\n[-] failed to create remote directory: %s", err)
            }
        } else {
            return err
        }
    }
    return nil
}