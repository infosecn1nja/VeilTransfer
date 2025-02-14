package handlers

import (
    "fmt"
    "io"
    "io/fs"
    "log"
    "net"
    "os"

    "github.com/goftp/server"
    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

// ========================
// ✅ FTP DRIVER & SERVER
// ========================
type FTPDriver struct {
    Username string
    Password string
    RootDir  string
}

func (d *FTPDriver) Init(conn *server.Conn) {}

func (d *FTPDriver) NewDriver() (server.Driver, error) {
    return d, nil
}

func (d *FTPDriver) CheckPasswd(user, pass string) (bool, error) {
    if user == d.Username && pass == d.Password {
        return true, nil
    }
    return false, fmt.Errorf("authentication failed")
}

func (d *FTPDriver) ChangeDir(path string) error {
    if _, err := os.Stat(path); err != nil {
        return err
    }
    return nil
}

func (d *FTPDriver) Stat(path string) (server.FileInfo, error) {
    info, err := os.Stat(path)
    if err != nil {
        return nil, err
    }
    return &fileInfo{info}, nil
}

func (d *FTPDriver) ListDir(path string, callback func(server.FileInfo) error) error {
    entries, err := os.ReadDir(path)
    if err != nil {
        return err
    }
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue
        }
        callback(&fileInfo{info})
    }
    return nil
}

func (d *FTPDriver) DeleteDir(path string) error {
    return os.Remove(path)
}

func (d *FTPDriver) DeleteFile(path string) error {
    return os.Remove(path)
}

func (d *FTPDriver) Rename(fromPath, toPath string) error {
    return os.Rename(fromPath, toPath)
}

func (d *FTPDriver) MakeDir(path string) error {
    return os.MkdirAll(path, 0755)
}

func (d *FTPDriver) GetFile(path string, offset int64) (int64, io.ReadCloser, error) {
    file, err := os.Open(path)
    if err != nil {
        return 0, nil, err
    }
    file.Seek(offset, 0)
    info, err := file.Stat()
    if err != nil {
        return 0, nil, err
    }
    return info.Size(), file, nil
}

func (d *FTPDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {
    flag := os.O_WRONLY | os.O_CREATE
    if appendData {
        flag |= os.O_APPEND
    } else {
        flag |= os.O_TRUNC
    }

    file, err := os.OpenFile(destPath, flag, 0644)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    return io.Copy(file, data)
}

func StartFTPServer(port int, driver *FTPDriver) {
    opts := &server.ServerOpts{
        Factory: driver,
        Port:    port,
        Auth:    driver,
    }

    ftpServer := server.NewServer(opts)
    log.Printf("FTP server started on port %d", port)

    if err := ftpServer.ListenAndServe(); err != nil {
        log.Fatalf("Failed to start FTP server: %v", err)
    }
}

// ========================
// ✅ SFTP SERVER HANDLER
// ========================
func StartSFTPServer(port int, username, password, keyPath string) {
    config := &ssh.ServerConfig{
        PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            if c.User() == username && string(pass) == password {
                return nil, nil
            }
            return nil, fmt.Errorf("password rejected for %q", c.User())
        },
    }

    privateBytes, err := os.ReadFile(keyPath)
    if err != nil {
        log.Fatalf("Failed to load private key: %v", err)
    }

    private, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil {
        log.Fatalf("Failed to parse private key: %v", err)
    }

    config.AddHostKey(private)

    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        log.Fatalf("Failed to listen for connection: %v", err)
    }
    log.Printf("SFTP server started on port %d", port)

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Failed to accept connection: %v", err)
            continue
        }

        go func() {
            sshConn, chans, _, err := ssh.NewServerConn(conn, config)
            if err != nil {
                log.Printf("Failed to handshake: %v", err)
                return
            }

            log.Printf("New SFTP connection from %s", sshConn.RemoteAddr())

            for newChannel := range chans {
                if newChannel.ChannelType() != "session" {
                    newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
                    continue
                }

                channel, requests, err := newChannel.Accept()
                if err != nil {
                    log.Printf("Could not accept channel: %v", err)
                    continue
                }

                go ssh.DiscardRequests(requests)

                sftpServer, err := sftp.NewServer(channel)
                if err != nil {
                    log.Printf("Failed to start SFTP subsystem: %v", err)
                    continue
                }

                if err := sftpServer.Serve(); err == io.EOF {
                    log.Printf("SFTP client disconnected.")
                } else if err != nil {
                    log.Printf("SFTP server error: %v", err)
                }
            }
        }()
    }
}

// ========================
// ✅ File Info Handler
// ========================
type fileInfo struct {
    os.FileInfo
}

func (f *fileInfo) Owner() string {
    return "user"
}

func (f *fileInfo) Group() string {
    return "group"
}

func (f *fileInfo) Mode() fs.FileMode {
    return f.FileInfo.Mode()
}
