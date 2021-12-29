package main

import (
	"client/callback"
	"fmt"
	"github.com/reujab/wallpaper"
	"log"
	"sync"
	"time"
)

var lock sync.Mutex

func SetWallpaper(url string) {
	lock.Lock()
	defer lock.Unlock()
	err := wallpaper.SetFromURL("http://" + url)
	if err != nil {
		log.Fatal(err)
	}
}

func checkForUpdates(url string, cb chan bool, ticker *time.Ticker) {
	SetWallpaper(url)
	for {
		select {
		case <-ticker.C:
			fmt.Println("30 min interval passed. Checking for updates...")
			SetWallpaper(url)
		case <-cb:
			fmt.Println("Callback received. Checking for updates...")
			SetWallpaper(url)
		}

	}
}

func main() {
	//10.10.0.209
	url := "127.0.0.1:10000"
	cb := make(chan bool)
	t := time.NewTicker(30 * time.Minute)
	go callback.CheckNastyCallback(cb)
	checkForUpdates(url, cb, t)
}
