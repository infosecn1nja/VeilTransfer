package cmd

import (
    "flag"
    "fmt"
    "os"
    "strings"
    "client/transfer"
    "client/archive"
    "client/generator"
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
        fmt.Println("You must specify a command: transfer, generate-fake or create-zip")
        return
    }

    command := os.Args[1]

    switch command {
    case "transfer":

        if len(os.Args) < 3 {
            fmt.Println("Available transfer methods:")
            fmt.Println("github, icmp, dropbox, sftp, ftp, webdav, mega, webhook, pastebin, telegram, doh, quic")
            return
        }

        method := os.Args[2]
        transferFlags := flag.NewFlagSet("transfer", flag.ExitOnError)

        var apiKey, username, password, server, remoteDir, url, telegramAPI, key, dnsServer, repo, localPath, include string
        var channelID int64
        var interval int

        switch method {
        case "dropbox":
            transferFlags.StringVar(&apiKey, "apiKey", "", "API key for Dropbox")
            transferFlags.StringVar(&remoteDir, "remoteDir", "", "Path to the remote directory")
        case "pastebin":
            transferFlags.StringVar(&apiKey, "apiKey", "", "API key for Pastebin")
        case "github":
            transferFlags.StringVar(&apiKey, "apiKey", "", "API key for Github")
            transferFlags.StringVar(&repo, "repo", "", "Repository name for GitHub (yourusername/repository)")
        case "sftp", "ftp", "webdav", "mega":
            transferFlags.StringVar(&username, "username", "", "Username for authentication")
            transferFlags.StringVar(&password, "password", "", "Password for authentication")
            transferFlags.StringVar(&server, "server", "", "Server address with port")
            transferFlags.StringVar(&remoteDir, "remoteDir", "", "Path to the remote directory")
        case "webhook":
            transferFlags.StringVar(&url, "url", "", "Webhook URL")
        case "telegram":
            transferFlags.StringVar(&telegramAPI, "telegramAPI", "", "Telegram bot API token")
            transferFlags.Int64Var(&channelID, "channelID", 0, "Telegram channel ID")
        case "doh":
            transferFlags.StringVar(&key, "key", "", "AES encryption key")
            transferFlags.StringVar(&dnsServer, "dnsServer", "", "Domain server")
        case "quic":
            transferFlags.StringVar(&server, "server", "", "Server address with port")
        case "icmp":
            transferFlags.StringVar(&server, "server", "", "Server address")
        }

        transferFlags.StringVar(&localPath, "localPath", "", "Path to the local file or directory")
        transferFlags.StringVar(&include, "include", "", "Comma-separated list of file patterns to include (e.g \"*.txt,*.csv,*.docx\")")
        transferFlags.IntVar(&interval, "interval", 0, "Interval in minutes between file uploads (default: 0 - no delay)")

        if err := transferFlags.Parse(os.Args[3:]); err != nil {
            fmt.Println("Error parsing flags:", err)
            return
        }

        var err error
        switch method {
        case "dropbox":
            err = transfer.UploadDropbox(apiKey, localPath, remoteDir, strings.Split(include, ","), interval)
        case "github":
            err = transfer.UploadGithub(apiKey, localPath, repo, strings.Split(include, ","), interval)
        case "sftp":
            err = transfer.UploadSFTP(username, password, server, localPath, remoteDir, "", strings.Split(include, ","), interval)
        case "ftp":
            err = transfer.UploadFTP(username, password, server, localPath, remoteDir, strings.Split(include, ","), interval)
        case "icmp":
            err = transfer.UploadICMP(server, localPath, strings.Split(include, ","), interval)
        case "webdav":
            err = transfer.UploadWebDAV(username, password, server, localPath, remoteDir, strings.Split(include, ","), interval)
        case "mega":
            err = transfer.UploadMega(username, password, localPath, remoteDir, strings.Split(include, ","), interval)
        case "webhook":
            err = transfer.UploadWebhook(localPath, url, strings.Split(include, ","), interval)
        case "pastebin":
            err = transfer.UploadPastebin(apiKey, localPath, strings.Split(include, ","), interval)
        case "telegram":
            bot, botErr := tgbotapi.NewBotAPI(telegramAPI)
            if botErr != nil {
                fmt.Printf("[-] Error during %s upload: %s\n", method, botErr)
                return
            }
            err = transfer.UploadTelegram(bot, localPath, channelID, strings.Split(include, ","), interval)
        case "doh":
            err = transfer.UploadDOH(key, dnsServer, localPath, strings.Split(include, ","), interval)
        case "quic":
            err = transfer.UploadQUIC(server, localPath, strings.Split(include, ","), interval)
        default:
            fmt.Println("[-] Invalid method specified.")
            return
        }

        if err != nil {
            fmt.Printf("[-] Error during %s upload: %s\n", method, err)
        } else {
            fmt.Printf("\n[*] File(s) uploaded successfully via %s\n", method)
        }

    case "generate-fake":
        fakeDataFlags := flag.NewFlagSet("generate-fake", flag.ExitOnError)
        count := fakeDataFlags.Int("count", 0, "Total number of entries to generate")
        generateKTP := fakeDataFlags.Bool("ktp", false, "Generate KTP data")
        generateSSN := fakeDataFlags.Bool("ssn", false, "Generate SSN data")
        generateCCN := fakeDataFlags.Bool("ccn", false, "Generate credit card data")
        generateMR := fakeDataFlags.Bool("medical-record", false, "Generate medical record data")
        language := fakeDataFlags.String("language", "en", "Set language for medical record data (id or en)")

        fakeDataFlags.Parse(os.Args[2:])

        if *count <= 0 {
            fmt.Println("Error: The 'count' must be greater than 0.")
            os.Exit(1)
        }

        if !*generateKTP && !*generateSSN && !*generateCCN && !*generateMR {
            fmt.Println("Error: You must choose at least one of the following: -ktp, -ssn, -ccn, -medical-record.")
            os.Exit(1)
        }

        if *generateMR {
            if *language != "id" && *language != "en" {
                fmt.Println("Error: Invalid language. Use '-language=id' or '-language=en'.")
                os.Exit(1)
            }
            generator.GenerateMedicalRecords(*count, *language)
        }

        if *generateKTP {
            generator.GenerateKTPs(*count)
        }

        if *generateSSN {
            generator.GenerateSSNs(*count)
        }

        if *generateCCN {
            generator.GenerateCreditCards(*count)
        }

    case "create-zip":
        createZipFlags := flag.NewFlagSet("create-zip", flag.ExitOnError)
        localPath := createZipFlags.String("localPath", "", "Path to the local file or directory")
        outputPath := createZipFlags.String("outputPath", "", "Output path for zip file")
        splitSize := createZipFlags.Int64("splitSize", 0, "Size in bytes to split the zip file")

        createZipFlags.Parse(os.Args[2:])

        if *localPath == "" || *outputPath == "" {
            fmt.Println("[-] Both localPath and outputPath must be provided for 'create-zip' command")
            return
        }

        err := archive.CreateZip(*localPath, *outputPath, *splitSize)
        if err != nil {
            fmt.Printf("[-] Error during zip creation: %s\n", err)
        } else {
            fmt.Printf("[*] Zip files created successfully with base name %s\n", *outputPath)
        }

    default:
        fmt.Printf("Invalid command specified: %s. Choose between 'transfer', 'generate-fake' and 'create-zip'.\n", command)
    }
}
