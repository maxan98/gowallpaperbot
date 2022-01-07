package clients

import (
	"github.com/onrik/gorm-logrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	IP       string
	LastSeen time.Time
}

var DB *gorm.DB

func Init() {
	log.Info("DB init and migration started (clients)")
	var err error
	hd, _ := os.UserHomeDir()
	dbpath := filepath.Join(hd, ".wppr", "clients.db")
	_, errc := os.Create(dbpath)
	if errc != nil {
		panic("failed to create database")
	}
	DB, err = gorm.Open(sqlite.Open(dbpath), &gorm.Config{Logger: gorm_logrus.New()})
	if err != nil {
		panic("failed to connect database")
	}

	// Миграция схем
	DB.AutoMigrate(&Client{})

}
func UpdateLastSeen(ip string) {
	var client Client
	DB.Where("ip = ?", ip).First(&client)
	client.LastSeen = time.Now()
	DB.Save(&client)
}
func GetAllClients() []Client {
	var clients []Client
	DB.Find(&clients)
	return clients
}
