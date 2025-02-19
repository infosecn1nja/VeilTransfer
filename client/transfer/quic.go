package transfer

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"client/utils"
	"github.com/quic-go/quic-go"
)

// UploadQUIC uploads files over QUIC with optional scheduling
func UploadQUIC(serverAddr, rootDir string, includePatterns []string, scheduleInterval int) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic"},
	}

	ctx := context.Background()
	conn, err := quic.DialAddr(ctx, serverAddr, tlsConfig, nil) // Use quic.Connection instead of quic.Session
	if err != nil {
		return fmt.Errorf("\n[-] QUIC dial error: %s", err)
	}
	defer conn.CloseWithError(0, "done")

	// Use WalkAndUpload for recursive directory traversal & scheduling
	return utils.WalkAndUpload(rootDir, "", includePatterns, func(localFilePath, remoteFilePath string) error {
		fileInfo, err := os.Stat(localFilePath)
		if err != nil {
			return fmt.Errorf("\n[-] Failed to stat local file: %s", err)
		}

		if fileInfo.IsDir() {
			fmt.Printf("\n[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
			return nil
		}

		err = uploadFileQUIC(ctx, conn, localFilePath, remoteFilePath, fileInfo)
		if err != nil {
			return fmt.Errorf("\n[-] Error uploading file: %s", err)
		}

		return nil
	}, scheduleInterval)
}

// uploadFileQUIC uploads a single file over QUIC
func uploadFileQUIC(ctx context.Context, conn quic.Connection, localFilePath, remoteFilePath string, fileInfo os.FileInfo) error {
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return fmt.Errorf("\n[-] Stream open error: %s", err)
	}
	defer stream.Close()

	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("\n[-] File open error: %s", err)
	}
	defer file.Close()

	relativePath := strings.ReplaceAll(remoteFilePath, "\\", "")
	nameBytes := []byte(relativePath)

	// Send filename length
	if err := binary.Write(stream, binary.LittleEndian, uint16(len(nameBytes))); err != nil {
		return err
	}
	// Send filename
	if _, err := stream.Write(nameBytes); err != nil {
		return err
	}
	// Send file size
	fileSize := fileInfo.Size()
	if err := binary.Write(stream, binary.LittleEndian, fileSize); err != nil {
		return err
	}

	// Progress tracking
	progress := make(chan int64, 1)
	go func() {
		for sent := range progress {
			percentage := float64(sent) / float64(fileSize) * 100
			fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
		}
	}()

	// Upload file in chunks
	buffer := make([]byte, 1*1024*1024)
	var totalSent int64

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := stream.Write(buffer[:n]); err != nil {
			return err
		}

		totalSent += int64(n)
		progress <- totalSent
	}

	close(progress)

	// Receive ACK
	ack := make([]byte, 3)
	if _, err := io.ReadFull(stream, ack); err != nil {
		return fmt.Errorf("\n[-] Error receiving ACK: %s", err)
	}
	if string(ack) != "ACK" {
		return fmt.Errorf("\n[-] Invalid ACK received")
	}

	return nil
}
