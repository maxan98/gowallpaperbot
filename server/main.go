package main

import (
	clients "client/clients"
	"client/settings"
	"client/share"
	"client/wallpaper"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func downloadFile(filepath string, url string, res chan bool) {
	settings := settings.GetInstance()
	// Create the file
	err := os.MkdirAll(settings.GetSettings().GetFilePath(), 0775)
	if err != nil {
		log.Panic("Failed to create file structure")
	}
	fmt.Println("saving to", settings.GetSettings().GetFilePath(), filepath)
	out, err := os.Create(settings.GetSettings().GetFilePath() + filepath)
	if err != nil {
		res <- false
		return
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		res <- false
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		res <- false
		return
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		res <- false
		return
	}
	res <- true

}

func processRequest(update tgbotapi.Update, ctx context.Context, log *zap.SugaredLogger, token string, bot *tgbotapi.BotAPI) {
	var msg tgbotapi.MessageConfig
	settings := settings.GetInstance()
	log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
	if update.Message.Photo != nil {
		pic := update.Message.Photo[len(update.Message.Photo)-1]
		log.Info(pic.FileID)
		file := &tgbotapi.File{FileID: pic.FileID}

		directURL, err := bot.GetFileDirectURL(file.FileID)
		if err != nil {
			log.Info("Unable to get file path from telegram server")
		}
		ext := filepath.Ext(directURL)
		if len(directURL) != 0 {
			reschan := make(chan bool)
			go downloadFile(file.FileID+ext, directURL, reschan)
			for {
				select {
				case <-ctx.Done():
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Error: Context Deadline Exceeded")
					break
				case res := <-reschan:
					if res == true {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Cool! Thanks for making my day better. I will update the wallpaper ASAP")
						img := wallpaper.GetPicture()
						err = img.SetPicture(settings.GetSettings().GetFilePath() + file.FileID + ext)
						if err != nil {
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Unexpected error when saving image"+err.Error())
						}

						cl := clients.GetInstance()

						for _, i := range cl.GetClients() {
							log.Info("Trying to ping the clients callback URL at ", i)
							cburl := fmt.Sprintf("http://%s:10001", i)
							_, err := http.Get(cburl)
							if err != nil {
								log.Info(err)
							}
						}

						break
					} else {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Unexpected error")
						break
					}
				}
				break
			}

		}
	} else {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Looks like there is no picture in your message. If you just want to chat - reach out to Max directly")
	}

	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		log.Info("error in sending response")
	}
}

func main() {
	//delete
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found")
	}
	log.Info("Trying to get Token from env TOKEN..")
	var token string
	var filePath string
	var ok bool
	if token, ok = os.LookupEnv("TOKEN"); !ok {
		log.Panic("Bot token unset", token)
	}
	log.Info(token)
	if filePath, ok = os.LookupEnv("FILES_DIR"); !ok {
		log.Panic("File Path unset", token)
	} else {
		filePath = "./files"
	}
	log.Info(token)
	set := settings.GetInstance()
	set.SetToken(token)
	set.SetFilePath(filePath)
	bot, err := tgbotapi.NewBotAPI(set.GetSettings().GetToken())
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go share.HandleRequests()

	for update := range updates {
		if update.Message != nil { // If we got a message
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			go processRequest(update, ctx, log, token, bot)
		}
	}
}
