package main

import (
	"client/callback"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/reujab/wallpaper"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var lock sync.Mutex

type Auth struct {
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
}

var Creds *Auth

func Init() {
	Creds = &Auth{}
	f, err := os.Open("./clientAuth.txt")
	if err != nil {
		panic("Creds not set")
	}
	d := yaml.NewDecoder(f)
	if err = d.Decode(&Creds); err != nil {
		panic("Unable to decode auth file")
	}
}

func downloadImage(url string) (string, error) {

	client := &http.Client{
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest("GET", "https://"+url, nil)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	req.SetBasicAuth(Creds.Username, Creds.Password)
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", errors.New("non-200 status code")
	}
	defer res.Body.Close()

	file, err := os.Create(filepath.Join(os.TempDir(), "wallpaper"))
	if err != nil {
		return "", err
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return "", err
	}

	err = file.Close()
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func SetWallpaper(url string) {
	lock.Lock()
	defer lock.Unlock()
	f, err := downloadImage(url)
	if err != nil {
		log.Print(err.Error())
	}
	err = wallpaper.SetFromFile(f)
	if err != nil {
		log.Print(err)
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
	Init()
	//10.10.0.209
	url := "10.10.0.209:10000"
	cb := make(chan bool)
	t := time.NewTicker(30 * time.Minute)
	go callback.CheckNastyCallback(cb)
	checkForUpdates(url, cb, t)
}
