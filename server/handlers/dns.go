package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base32"
	"fmt"
	"github.com/miekg/dns"
	"os"
	"sort"
	"strings"
)

type FileSession struct {
	fileData       []byte
	fileNameChunks map[int]string
}

var (
	sessions = make(map[string]*FileSession)
	key      []byte
	saveFolder string
)

func decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return data, nil
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)

	for _, q := range r.Question {
		if q.Qtype == dns.TypeTXT {
			queryName := strings.TrimSuffix(q.Name, ".")
			labels := strings.Split(queryName, ".")

			if len(labels) < 2 {
				continue
			}
			sessionID := labels[1]

			if _, exists := sessions[sessionID]; !exists {
				sessions[sessionID] = &FileSession{
					fileNameChunks: make(map[int]string),
				}
			}
			session := sessions[sessionID]

			if strings.HasPrefix(queryName, "FILENAME") {
				index := 0
				fmt.Sscanf(labels[0][8:], "%d", &index)
				session.fileNameChunks[index] = labels[2]
				fmt.Printf("Received filename chunk [%d]: %s (Session: %s)\n", index, labels[2], sessionID)
			} else if strings.HasPrefix(queryName, "END") {
				fmt.Println("File transfer completed for session:", sessionID)

				if len(session.fileNameChunks) == 0 {
					fmt.Println("Error: No filename chunks received.")
					continue
				}

				var keys []int
				for k := range session.fileNameChunks {
					keys = append(keys, k)
				}
				sort.Ints(keys)

				var base32FileName string
				for _, k := range keys {
					base32FileName += session.fileNameChunks[k]
				}

				decodedFileName, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(base32FileName)
				if err != nil {
					fmt.Println("Error decoding filename:", err)
					continue
				}
				fileName := string(decodedFileName)
				fmt.Println("Receiving file with original name:", fileName)

				decryptedData, err := decrypt(session.fileData, key)
				if err != nil {
					fmt.Println("Error decrypting file:", err)
					continue
				}

				filePath := fmt.Sprintf("%s/%s", saveFolder, fileName)
				file, err := os.Create(filePath)
				if err != nil {
					fmt.Println("Error creating file:", err)
					continue
				}

				_, err = file.Write(decryptedData)
				if err != nil {
					fmt.Println("Error writing to file:", err)
				} else {
					fmt.Printf("File saved successfully at %s\n", filePath)
				}
				file.Close()

				delete(sessions, sessionID)
				fmt.Println("Ready for the next file transfer.")
			} else {
				encodedChunk := labels[0]
				decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(encodedChunk)
				if err != nil {
					fmt.Println("Decoding error:", err)
					continue
				}
				session.fileData = append(session.fileData, decoded...)
			}

			msg.Answer = append(msg.Answer, &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{"ACK"},
			})
		}
	}

	w.WriteMsg(msg)
}

// Modified StartDNSServer to accept key and saveFolder as arguments
func StartDNSServer(aesKey []byte, folder string) {
	key = aesKey
	saveFolder = folder

	srv := &dns.Server{Addr: ":53", Net: "udp"}
	dns.HandleFunc(".", handleDNSRequest)

	fmt.Println("DNS Server is running on port 53")
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
