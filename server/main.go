package main

import (
	"client/chats"
	clients "client/clients"
	"client/settings"
	"client/share"
	"client/wallpaper"
	"context"
	"fmt"
	"github.com/fogleman/gg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/freetype/truetype"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"image/jpeg"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func ApplyLabel(img image.Image, text string) image.Image {
	dc := gg.NewContextForImage(img)
	font, _ := truetype.Parse(goregular.TTF)
	ff := truetype.NewFace(font, &truetype.Options{Size: 124})
	dc.SetFontFace(ff)
	words := dc.WordWrap(text, 3000.)
	y := 300.
	dc.SetRGBA255(87, 166, 198, 35)

	dc.DrawRoundedRectangle(130., 150., 3000., float64(len(words)*150.)+50, 50)
	dc.Fill()
	for _, i := range words {
		dc.SetRGBA255(31, 74, 911, 205)
		dc.DrawStringAnchored(i, 150., y, 0., 0.)
		y += 150
	}
	i := dc.Image()
	return i
}

func downloadFile(filepath string, url string, res chan int, text string) {
	set := settings.GetInstance()
	// Create the file
	err := os.MkdirAll(set.GetSettings().GetFilePath(), 0775)
	if err != nil {
		log.Panic("Failed to create file structure")
	}
	log.Info("saving to", set.GetSettings().GetFilePath(), filepath)
	out, err := os.Create(set.GetSettings().GetFilePath() + filepath)
	if err != nil {
		res <- 0
		return
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		res <- 0
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		res <- 0
		return
	}
	if text != "" {
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Info("Unable to apply the label")

		} else {

		}
		imgLabeled := ApplyLabel(img, text)
		opt := jpeg.Options{
			Quality: 90,
		}
		err = jpeg.Encode(out, imgLabeled, &opt)
		if err != nil {
			res <- 0
			return
		}else {
			res <- 2
			return
		}
	} else {
		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			res <- 0
			return
		}
	}

	res <- 1

}
func cleanupFiles() {
	set := settings.GetInstance()
	img := wallpaper.GetPicture()
	_, imgpath, _ := img.GetPicture()
	files, err := os.ReadDir(set.GetSettings().GetFilePath())
	if err != nil {
		log.Info("Error while cleanin up files")
	}
	for _, i := range files {
		if !i.IsDir() && set.GetSettings().GetFilePath()+i.Name() != imgpath {
			log.Info("iname", set.GetSettings().GetFilePath()+i.Name(), "filepath", imgpath)
			err = os.Remove(set.GetSettings().GetFilePath() + i.Name())
			if err != nil {
				log.Info("Error while deleting old image", err.Error())
			}
		}
	}
}
func processRequest(update tgbotapi.Update, ctx context.Context, bot *tgbotapi.BotAPI) {
	var msg tgbotapi.MessageConfig
	settings := settings.GetInstance()
	log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
	if update.Message.Photo != nil || update.Message.Document != nil {
		var file *tgbotapi.File
		if len(update.Message.Photo) != 0 {
			pic := update.Message.Photo[len(update.Message.Photo)-1]

			log.Info(pic.FileID)
			file = &tgbotapi.File{FileID: pic.FileID}
		} else if update.Message.Document != nil {
			file = &tgbotapi.File{FileID: update.Message.Document.FileID}
		}
		directURL, err := bot.GetFileDirectURL(file.FileID)
		if err != nil {
			log.Info("Unable to get file path from telegram server")
		}
		ext := filepath.Ext(directURL)
		if len(directURL) != 0 {
			reschan := make(chan int)
			go downloadFile(file.FileID+ext, directURL, reschan, update.Message.Text)

			select {
			case <-ctx.Done():
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Error: Context Deadline Exceeded")
				break
			case res := <-reschan:
				if res == 1 || res == 2 {
					if res == 1{
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Cool! Thanks for making my day better. I will update the wallpaper ASAP")
					}
					if res == 2{
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Cool! Thanks for making my day better. I will update the wallpaper ASAP. I have also noticed" +
							" that you wanted to insert some text into the wallpaper - will be done :) ")
					}
					img := wallpaper.GetPicture()
					err = img.SetPicture(settings.GetSettings().GetFilePath() + file.FileID + ext)
					if err != nil {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Unexpected error when saving image"+err.Error())
					}

					cl := clients.GetInstance()
					cleanupFiles()
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

		}
	} else {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Looks like there is no picture in your message. If you just want to chat - reach out to Max directly")
	}

	msg.ReplyToMessageID = update.Message.MessageID
	log.Infof("Finishing processing request from %s, sending resp %s", update.Message.From.FirstName, msg.Text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Info("error in sending response")
	}
}

func sendReminder(ticker *time.Ticker, bot *tgbotapi.BotAPI) {
	log.Info("Starting reminder goroutine")
	s3 := rand.NewSource(time.Now().UnixNano())
	r3 := rand.New(s3)
	for {
		select {
		case <-ticker.C:
			newRandom := r3.Intn(10)
			if newRandom > 5 && time.Now().Local().Hour() <= 23 && time.Now().Local().Hour() >= 9 {
				chatList := chats.GetInstance()
				log.Info("random decided to update ", newRandom)
				log.Info(chatList)
				ch := chatList.GetClients()
				for _, i := range ch {
					msg := tgbotapi.NewMessage(i, "Hey! It is been a while since you last updated the wallpaper. Please find some time to do so :)")
					_, err := bot.Send(msg)
					if err != nil {
						log.Info("error in sending response")
					}
				}

			} else {
				log.Info("Random decided not to update right now")
			}
		}
	}
}

func main() {

	fmt.Println("Loading config..")
	configPath, err := settings.ParseFlags()
	if err != nil {
		panic("no configpath " + err.Error())
	}
	set, err := settings.NewConfig(configPath)
	if err != nil {
		panic("Unable to parse config " + err.Error())
	}

	//delete
	lfile, err := os.Create(set.LogFile)
	if err != nil {
		panic(err.Error())
	}
	mw := io.MultiWriter(os.Stdout, lfile)
	log.SetOutput(mw)

	bot, err := tgbotapi.NewBotAPI(set.GetSettings().GetToken())
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Infof("Authorized on account %s", bot.Self.UserName)
	tgbotapi.SetLogger(log.StandardLogger())
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	err = os.MkdirAll(set.GetSettings().GetFilePath(), 0775)
	if err != nil {
		log.Panic("Failed to create file structure")
	}
	go share.HandleRequests()
	tick := time.NewTicker(213 * time.Minute)
	go sendReminder(tick, bot)
	for update := range updates {
		if update.Message != nil { // If we got a message
			for _, i := range set.AllowedIDs {
				if update.Message.From.ID == i {
					chatList := chats.GetInstance()
					chatList.AppendClient(update.Message.Chat.ID)

					ctx, _ := context.WithTimeout(context.Background(), 25*time.Second)

					go processRequest(update, ctx, bot)
				}
			}
		}
	}
}
