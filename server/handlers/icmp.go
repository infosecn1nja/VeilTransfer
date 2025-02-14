package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	xorKey = 0xAA // XOR encryption key
)

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

func handleICMPPacket(conn net.PacketConn, buf []byte, n int, addr net.Addr, file *os.File, filename *string, destDir string) *os.File {
	decryptedData := xorEncryptDecrypt(buf[8:n])

	if string(decryptedData) == "EOF" {
		if file != nil {
			file.Close()
			fmt.Printf("File %s received completely from %s\n", *filename, addr)
		}
		*filename = ""
		return nil
	}

	if *filename == "" {
		parts := strings.SplitN(string(decryptedData), "|", 2)
		if len(parts) == 2 {
			*filename = parts[0]
			filePath := filepath.Join(destDir, *filename)
			var err error
			file, err = os.Create(filePath)
			if err != nil {
				log.Fatalf("Failed to create file: %v", err)
			}
			file.Write([]byte(parts[1]))
			fmt.Printf("Created file: %s from %s\n", filePath, addr)
			return file
		}
	}

	if file != nil {
		file.Write(decryptedData)
		fmt.Printf("Received data chunk from %s\n", addr)
	}
	return file
}

func StartICMPServer(destDir string) {
	if destDir == "" {
		log.Fatal("Destination directory is required")
	}

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err = os.Mkdir(destDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create destination directory: %v", err)
		}
		fmt.Printf("Created destination directory: %s\n", destDir)
	} else {
		fmt.Printf("Using existing destination directory: %s\n", destDir)
	}

	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatalf("Failed to start ICMP server: %v", err)
	}
	defer conn.Close()

	fmt.Println("ICMP server started and listening on 0.0.0.0")

	buf := make([]byte, 1500)
	var file *os.File
	var filename string

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("Error reading packet: %v", err)
			continue
		}
		file = handleICMPPacket(conn, buf, n, addr, file, &filename, destDir)
	}
}
