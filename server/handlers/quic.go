package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/quic-go/quic-go"
)

func StartQUICServer(address, certFile, keyFile, outputDir string) {
	tlsConfig := generateTLSConfig(certFile, keyFile)

	listener, err := quic.ListenAddr(address, tlsConfig, &quic.Config{
		MaxIncomingStreams: 1000,
	})
	if err != nil {
		fmt.Println("Failed to start QUIC listener:", err)
		os.Exit(1)
	}
	fmt.Println("QUIC server listening on", address)

	for {
		session, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Println("Failed to accept QUIC session:", err)
			continue
		}
		go handleSession(session, outputDir)
	}
}


func generateTLSConfig(certFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Println("Error loading TLS certificates:", err)
		os.Exit(1)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic"},
	}
}

func handleSession(session quic.Connection, outputDir string) {
	ctx := context.Background()
	for {
		stream, err := session.AcceptStream(ctx)
		if err != nil {
			if session.Context().Err() != nil {
				fmt.Println("\nSession closed gracefully.")
				return
			}
			return
		}
		go func(stream quic.Stream) {
			defer stream.Close()

			buffer := make([]byte, 4)
			n, err := stream.Read(buffer)
			if err == nil && string(buffer[:n]) == "DONE" {
				fmt.Println("\nTransfer completed. Session closed gracefully.")
				session.CloseWithError(0, "done")
				return
			}

			handleStream(stream, buffer[:n], outputDir)
		}(stream)
	}
}

func handleStream(stream quic.Stream, initialData []byte, outputDir string) {
	reader := io.MultiReader(
		bytes.NewReader(initialData),
		stream,
	)

	var nameLen uint16
	if err := binary.Read(reader, binary.LittleEndian, &nameLen); err != nil {
		fmt.Println("Error reading filename length:", err)
		return
	}

	nameBuf := make([]byte, nameLen)
	if _, err := io.ReadFull(reader, nameBuf); err != nil {
		fmt.Println("Error reading filename:", err)
		return
	}
	relativePath := string(nameBuf)

	var fileSize int64
	if err := binary.Read(reader, binary.LittleEndian, &fileSize); err != nil {
		fmt.Println("Error reading file size:", err)
		return
	}

	outputPath := filepath.Join(outputDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		fmt.Println("Error creating directories:", err)
		return
	}

	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("File creation error:", err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 1*1024*1024)
	var totalReceived int64

	fmt.Printf("Receiving '%s' (%.2f MB)...\n", relativePath, float64(fileSize)/(1024*1024))
	for totalReceived < fileSize {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("File read error:", err)
			return
		}
		if n == 0 {
			break
		}
		if _, err := file.Write(buffer[:n]); err != nil {
			fmt.Println("File write error:", err)
			return
		}
		totalReceived += int64(n)
		printProgress(totalReceived, fileSize)
	}

	fmt.Printf("\nFile '%s' received successfully.\n", relativePath)
	if _, err := stream.Write([]byte("ACK")); err != nil {
		fmt.Println("Error sending ACK:", err)
		return
	}
}

func printProgress(received, total int64) {
	percent := float64(received) / float64(total) * 100
	fmt.Printf("\rProgress: %.2f%%", percent)
}
