package handlers

import (
    "encoding/binary"
    "fmt"
    "io"
    "io/fs"
    "log"
    "net"
    "os"
    "path/filepath"

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

func (d *FTPDriver) abs(p string) string {
    // Optional: enforce rootdir sandboxing for FTP
    if d.RootDir == "" {
        return p
    }
    clean := filepath.Clean("/" + p) // keep absolute-like
    return filepath.Join(d.RootDir, clean)
}

func (d *FTPDriver) ChangeDir(path string) error {
    path = d.abs(path)
    if _, err := os.Stat(path); err != nil {
        return err
    }
    return nil
}

func (d *FTPDriver) Stat(path string) (server.FileInfo, error) {
    path = d.abs(path)
    info, err := os.Stat(path)
    if err != nil {
        return nil, err
    }
    return &fileInfo{info}, nil
}

func (d *FTPDriver) ListDir(path string, callback func(server.FileInfo) error) error {
    path = d.abs(path)
    entries, err := os.ReadDir(path)
    if err != nil {
        return err
    }
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue
        }
        _ = callback(&fileInfo{info})
    }
    return nil
}

func (d *FTPDriver) DeleteDir(path string) error {
    path = d.abs(path)
    return os.Remove(path)
}

func (d *FTPDriver) DeleteFile(path string) error {
    path = d.abs(path)
    return os.Remove(path)
}

func (d *FTPDriver) Rename(fromPath, toPath string) error {
    fromPath = d.abs(fromPath)
    toPath = d.abs(toPath)
    return os.Rename(fromPath, toPath)
}

func (d *FTPDriver) MakeDir(path string) error {
    path = d.abs(path)
    return os.MkdirAll(path, 0755)
}

func (d *FTPDriver) GetFile(path string, offset int64) (int64, io.ReadCloser, error) {
    path = d.abs(path)
    file, err := os.Open(path)
    if err != nil {
        return 0, nil, err
    }
    _, _ = file.Seek(offset, 0)

    info, err := file.Stat()
    if err != nil {
        _ = file.Close()
        return 0, nil, err
    }

    return info.Size(), file, nil
}

func (d *FTPDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {
    destPath = d.abs(destPath)

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
//
// Fix utama:
// - Handle request "subsystem" dan reply true untuk "sftp"
// - Jangan discard channel requests sebelum diproses
//
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

        go handleSFTPConn(conn, config)
    }
}

func handleSFTPConn(conn net.Conn, config *ssh.ServerConfig) {
    defer conn.Close()

    sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
    if err != nil {
        log.Printf("Failed to handshake: %v", err)
        return
    }
    defer sshConn.Close()

    // Discard global requests (keep connection clean)
    go ssh.DiscardRequests(reqs)

    log.Printf("New SSH connection from %s (user=%s)", sshConn.RemoteAddr(), sshConn.User())

    for newChannel := range chans {
        if newChannel.ChannelType() != "session" {
            _ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
            continue
        }

        channel, requests, err := newChannel.Accept()
        if err != nil {
            log.Printf("Could not accept channel: %v", err)
            continue
        }

        go handleSFTPSession(channel, requests)
    }
}

func handleSFTPSession(ch ssh.Channel, in <-chan *ssh.Request) {
    defer ch.Close()

    for req := range in {
        switch req.Type {

        case "subsystem":
            // payload: uint32 len + string (ex: "sftp")
            name, ok := parseSSHString(req.Payload)
            if !ok {
                _ = req.Reply(false, nil)
                continue
            }

            if name != "sftp" {
                _ = req.Reply(false, nil)
                continue
            }

            // IMPORTANT: client butuh reply true supaya tidak "subsystem request failed"
            _ = req.Reply(true, nil)

            srv, err := sftp.NewServer(ch)
            if err != nil {
                log.Printf("Failed to start SFTP server: %v", err)
                return
            }

            log.Printf("SFTP subsystem started")

            if err := srv.Serve(); err == io.EOF {
                log.Printf("SFTP client disconnected")
            } else if err != nil {
                log.Printf("SFTP server error: %v", err)
            }
            return

        default:
            // optional: to be strict, reject others.
            // Some clients may send "env", "pty-req", "shell", "exec".
            // For SFTP-only server, we return false.
            _ = req.Reply(false, nil)
        }
    }
}

func parseSSHString(payload []byte) (string, bool) {
    if len(payload) < 4 {
        return "", false
    }
    n := binary.BigEndian.Uint32(payload[:4])
    if int(4+n) > len(payload) {
        return "", false
    }
    return string(payload[4 : 4+n]), true
}

// ========================
// ✅ File Info Handler
// ========================
type fileInfo struct {
    os.FileInfo
}

func (f *fileInfo) Owner() string { return "user" }
func (f *fileInfo) Group() string { return "group" }
func (f *fileInfo) Mode() fs.FileMode {
    return f.FileInfo.Mode()
}
