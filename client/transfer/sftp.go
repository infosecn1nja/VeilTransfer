package transfer

import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
    "strings"

    "client/utils"
    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

func UploadSFTP(username, password, server, localPath, remoteDir, privateKeyPath string, includePatterns []string, scheduleInterval int) error {
    var authMethods []ssh.AuthMethod

    if privateKeyPath != "" {
        key, err := ioutil.ReadFile(privateKeyPath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to read private key file: %s", err)
        }

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

    remoteDir = cleanRemote(remoteDir)

    // Ensure base remote dir exists
    if err := ensureRemoteDir(client, remoteDir); err != nil {
        return fmt.Errorf("\n[-] Failed to ensure remote directory: %s", err)
    }

    // WalkAndUpload: recursive directory traversal and scheduling
    return utils.WalkAndUpload(localPath, remoteDir, includePatterns, func(localFilePath, remoteFilePath string) error {
        localFilePath = strings.ReplaceAll(localFilePath, "\\", "/")
        remoteFilePath = cleanRemote(remoteFilePath)

        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("\n[-] Failed to stat local file: %s", err)
        }

        // If directory => create remote directory and return
        if fileInfo.IsDir() {
            if remoteFilePath == "" {
                remoteFilePath = remoteDir
            }
            if err := client.MkdirAll(remoteFilePath); err != nil {
                return fmt.Errorf("\n[-] Failed to create remote directory: %s", err)
            }
            return nil
        }

        remoteFilePath = resolveRemoteFilePath(client, remoteDir, localFilePath, remoteFilePath)

        fmt.Printf("[*] Uploading %s to %s\n", localFilePath, remoteFilePath)

        if err := uploadFileSFTP(client, localFilePath, remoteFilePath, fileInfo); err != nil {
            return fmt.Errorf("\n[-] Error uploading %s: %s", localFilePath, err)
        }

        return nil
    }, scheduleInterval)
}

func resolveRemoteFilePath(client *sftp.Client, remoteDir, localFilePath, remoteFilePath string) string {
    base := filepath.Base(localFilePath)

    // If empty, just use remoteDir/base
    if remoteFilePath == "" {
        return path.Join(remoteDir, base)
    }

    if strings.HasSuffix(remoteFilePath, "/") {
        return path.Join(remoteFilePath, base)
    }

    if remoteFilePath == remoteDir {
        return path.Join(remoteDir, base)
    }

    if st, err := client.Stat(remoteFilePath); err == nil && st.IsDir() {
        return path.Join(remoteFilePath, base)
    }

    return remoteFilePath
}

func uploadFileSFTP(client *sftp.Client, localFilePath, remoteFilePath string, fileInfo os.FileInfo) error {
    srcFile, err := os.Open(localFilePath)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to open source file: %s", err)
    }
    defer srcFile.Close()

    parent := path.Dir(remoteFilePath)
    if parent != "." && parent != "/" {
        _ = client.MkdirAll(parent)
    }

    dstFile, err := client.Create(remoteFilePath)
    if err != nil {
        return fmt.Errorf("\n[-] Failed to create destination file: %s", err)
    }
    defer dstFile.Close()

    buf := make([]byte, 1024*1024)
    var total int64
    totalSize := fileInfo.Size()

    for {
        n, rerr := srcFile.Read(buf)
        if rerr != nil && rerr != io.EOF {
            return fmt.Errorf("\n[-] Failed to read source file: %s", rerr)
        }
        if n == 0 {
            break
        }

        if _, werr := dstFile.Write(buf[:n]); werr != nil {
            return fmt.Errorf("\n[-] Failed to write to destination file: %s", werr)
        }

        total += int64(n)
        percentage := float64(total) / float64(totalSize) * 100
        fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
    }

    return nil
}

func ensureRemoteDir(client *sftp.Client, remoteDir string) error {
    remoteDir = cleanRemote(remoteDir)

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

func cleanRemote(p string) string {
    p = strings.TrimSpace(p)
    p = strings.ReplaceAll(p, "\\", "/")
    if len(p) > 1 {
        p = strings.TrimRight(p, "/")
    }
    if p == "" {
        return "/"
    }
    return p
}
