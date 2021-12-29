package wallpaper

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"image"
	"os"
	"sync"
)

type picture struct {
	filePath string
	data     image.Image
	format   string
	lock     sync.RWMutex
}

var instance *picture
var once sync.Once

func GetPicture() *picture {
	once.Do(func() {
		instance = new(picture)
	})
	return instance
}

func (p *picture) SetPicture(filepath string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	f, err := os.Open(filepath)
	if err != nil {
		log.Info(err.Error())
		return errors.New("Unable to get wallpaper")
	}
	defer f.Close()

	img, fmtName, err := image.Decode(f)
	if err != nil {
		return errors.New("Unable to decode wallpaper")
	}
	instance.data = img
	instance.format = fmtName
	instance.filePath = filepath

	return nil
}

func (p *picture) GetPicture() (image.Image, string, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	log.Info(p.format, p.filePath, "hey")
	return instance.data, instance.filePath, nil
}
