package cmd

import (
    "flag"
    "fmt"
    "strings"
    "os"
    "log"
    "VeilTransfer/transfer"
    "VeilTransfer/archive"
    "VeilTransfer/generator"
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Main() {
    fmt.Println(`
 _____     _ _ _____                 ___         
|  |  |___|_| |_   _|___ ___ ___ ___|  _|___ ___ 
|  |  | -_| | | | | |  _| .'|   |_ -|  _| -_|  _|
 \___/|___|_|_| |_| |_| |__,|_|_|___|_| |___|_|  
             v1.0 | by @infosecn1nja                              
    `)
    if len(os.Args) < 2 {
        fmt.Println("You must specify a command: transfer, generate-fake or create-zip")
        return
    }

    command := os.Args[1]

    switch command {
    case "transfer":
        transferFlags := flag.NewFlagSet("transfer", flag.ExitOnError)
        method := transferFlags.String("method", "", "Transfer method: sftp, ftp, webdav, mega, webhook, pastebin, github, telegram")
        username := transferFlags.String("username", "", "Username for authentication")
        password := transferFlags.String("password", "", "Password for authentication (leave empty if using a private key)")
        server := transferFlags.String("server", "", "Server address with port (not used for mega or webhook)")
        localPath := transferFlags.String("localPath", "", "Path to the local file or directory")
        remoteDir := transferFlags.String("remoteDir", "", "Path to the remote directory")
        privateKeyPath := transferFlags.String("privateKey", "", "Path to the SSH private key file (optional, only for 'sftp' method)")
        webhookURL := transferFlags.String("webhookURL", "", "Webhook URL (only for 'webhook' method)")
        apiKey := transferFlags.String("apiKey", "", "API key for Pastebin and Github")
        includePatterns := transferFlags.String("include", "", "Comma-separated list of file patterns to include (e.g. *.xlsx,*.xls,*.docx)")
        repo := transferFlags.String("repo", "", "GitHub repository (in format 'owner/repo')")
        telegramAPI := transferFlags.String("telegramAPI", "", "Telegram bot API token (only for 'telegram' method)")
        channelID := transferFlags.Int64("channelID", 0, "Telegram channel ID (only for 'telegram' method)")

        transferFlags.Parse(os.Args[2:])

        var err error

        var includePatternsList []string
        if *includePatterns != "" {
            includePatternsList = strings.Split(*includePatterns, ",")
        }

        if len(includePatternsList) == 0 {
            fmt.Println("[!] No include patterns provided; all files will be uploaded.")
        }

        switch *method {
        case "github":
            err = transfer.UploadGithub(*apiKey, *localPath, *repo, includePatternsList)
        case "sftp":
            err = transfer.UploadSFTP(*username, *password, *server, *localPath, *remoteDir, *privateKeyPath, includePatternsList)
        case "ftp":
            err = transfer.UploadFTP(*username, *password, *server, *localPath, *remoteDir, includePatternsList)
        case "webdav":
            if !strings.HasPrefix(*server, "http://") && !strings.HasPrefix(*server, "https://") {
                fmt.Println("[-] WebDAV server address must include 'http://' or 'https://'.")
                return
            }
            err = transfer.UploadWebDAV(*username, *password, *server, *localPath, *remoteDir, includePatternsList)
        case "mega":
            err = transfer.UploadMega(*username, *password, *localPath, *remoteDir, includePatternsList)
        case "webhook":
            err = transfer.UploadWebhook(*localPath, *webhookURL, includePatternsList)
        case "pastebin":
             err = transfer.UploadPastebin(*apiKey, *localPath, includePatternsList)
        case "telegram":
            if *telegramAPI == "" || *channelID == 0 {
                fmt.Println("[-] Telegram API token and channel ID must be provided for the 'telegram' method.")
                return
            }
            bot, err := tgbotapi.NewBotAPI(*telegramAPI)
            if err != nil {
                log.Panic(err)
            }
            err = transfer.UploadTelegram(bot, *localPath, *channelID, includePatternsList)
        default:
            fmt.Println("[-] Invalid method specified. Choose between 'scp', 'ftp', 'webdav', 'mega', 'webhook', or 'pastebin'.")
            return
        }

        if err != nil {
            fmt.Printf("[-] Error during %s upload %s\n", *method, err)
        } else {
            fmt.Printf("\n[*] File(s) uploaded successfully via %s\n", *method)
        }

    case "generate-fake":
        fakeDataFlags := flag.NewFlagSet("generate-fake", flag.ExitOnError)
        count := fakeDataFlags.Int("count", 0, "Total number of entries to generate")
        generateKTP := fakeDataFlags.Bool("ktp", false, "Generate KTP data")
        generateSSN := fakeDataFlags.Bool("ssn", false, "Generate SSN data")
        generateCCN := fakeDataFlags.Bool("ccn", false, "Generate credit card data")

        fakeDataFlags.Parse(os.Args[2:])

        // Validation
        if *count <= 0 {
            fmt.Println("Error: The 'count' must be greater than 0.")
            os.Exit(1)
        }

        if !*generateKTP && !*generateSSN && !*generateCCN {
            fmt.Println("Error: You must choose at least one of the following: -ktp, -ssn, -ccn.")
            os.Exit(1)
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
        outputPath := createZipFlags.String("outputPath", "", "Output path for zip file (only for 'create-zip' command)")
        splitSize := createZipFlags.Int64("splitSize", 0, "Size in bytes to split the zip file into (only for 'create-zip' command)")

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
