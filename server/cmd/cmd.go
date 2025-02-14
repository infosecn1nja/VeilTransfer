package cmd

import (
	"flag"
	"fmt"
	"os"
	"server/handlers"
)

func Main() {
    fmt.Println(`
 _____     _ _ _____                 ___         
|  |  |___|_| |_   _|___ ___ ___ ___|  _|___ ___ 
|  |  | -_| | | | | |  _| .'|   |_ -|  _| -_|  _|
 \___/|___|_|_| |_| |_| |__,|_|_|___|_| |___|_|  
             v2.0 | by @infosecn1nja                              
    `)	
	if len(os.Args) < 2 {
		fmt.Println("You must specify a command: [icmp|webdav|doh|quic|ftp|sftp] [options]")
		return
	}

	subCommand := os.Args[1]
	switch subCommand {
	case "webdav":
		webdavCmd := flag.NewFlagSet("webdav", flag.ExitOnError)
		useSSL := webdavCmd.Bool("ssl", false, "Enable SSL (HTTPS) mode")
		certFile := webdavCmd.String("cert", "server.crt", "Path to SSL certificate file")
		keyFile := webdavCmd.String("key", "server.key", "Path to SSL key file")
		port := webdavCmd.String("port", "8080", "Port to run the server on")
		destDir := webdavCmd.String("dest", "/tmp/", "Destination directory to serve")
		username := webdavCmd.String("username", "", "Username for basic authentication")
		password := webdavCmd.String("password", "", "Password for basic authentication")
		webdavCmd.Parse(os.Args[2:])

		if *username == "" || *password == "" {
			fmt.Println("Username and password are required for the WebDAV server")
		}

		handlers.StartWebdav(*useSSL, *certFile, *keyFile, *port, *destDir, *username, *password)
	case "doh":
		dohCmd := flag.NewFlagSet("doh", flag.ExitOnError)
		dnsKey := dohCmd.String("key", "", "AES key for DNS server decryption")
		dnsFolder := dohCmd.String("folder", "/tmp", "Folder to save files received via DNS")
		dohCmd.Parse(os.Args[2:])

		if *dnsKey != "" {
			handlers.StartDNSServer([]byte(*dnsKey), *dnsFolder)
		}
	case "icmp":
		icmpCmd := flag.NewFlagSet("icmp", flag.ExitOnError)
		icmpFolder := icmpCmd.String("folder", "/tmp", "Folder to save files received via DNS")
		icmpCmd.Parse(os.Args[2:])

		handlers.StartICMPServer(*icmpFolder)

	case "quic":
		quicCmd := flag.NewFlagSet("quic", flag.ExitOnError)
		quicPort := quicCmd.String("port", "443", "Port for QUIC server")
		quicFolder := quicCmd.String("folder", "/tmp", "Folder to save files received via QUIC")
		quicCert := quicCmd.String("cert", "server.crt", "Path to TLS certificate")
		quicKey := quicCmd.String("key", "server.key", "Path to TLS key")
		quicCmd.Parse(os.Args[2:])

		handlers.StartQUICServer(":"+*quicPort, *quicCert, *quicKey, *quicFolder)

	case "ftp":
		ftpCmd := flag.NewFlagSet("ftp", flag.ExitOnError)
		ftpPort := ftpCmd.Int("port", 21, "Port for FTP server")
		ftpUser := ftpCmd.String("username", "", "Username for FTP authentication")
		ftpPass := ftpCmd.String("password", "", "Password for FTP authentication")
		ftpCmd.Parse(os.Args[2:])

		driver := &handlers.FTPDriver{Username: *ftpUser, Password: *ftpPass}
		handlers.StartFTPServer(*ftpPort, driver)

	case "sftp":
		sftpCmd := flag.NewFlagSet("sftp", flag.ExitOnError)
		sftpPort := sftpCmd.Int("sport", 22, "Port for SFTP server")
		sftpUser := sftpCmd.String("username", "", "Username for SFTP authentication")
		sftpPass := sftpCmd.String("password", "", "Password for SFTP authentication")
		sftpKey := sftpCmd.String("key", "id_rsa", "Path to the private key for SFTP server")
		sftpCmd.Parse(os.Args[2:])

		handlers.StartSFTPServer(*sftpPort, *sftpUser, *sftpPass, *sftpKey)

	default:
		fmt.Println("Unknown command:", subCommand)
		fmt.Println("Invalid command specified: [icmp|webdav|doh|quic|ftp|sftp] [options]")
	}
}
