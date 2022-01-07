package users

import (
	gorm_logrus "github.com/onrik/gorm-logrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type User struct {
	UserID   int64 `gorm:"primarykey"`
	ChatID   int64
	UserName string
	LastSeen time.Time
}

var DB *gorm.DB

func Init() {
	log.Info("DB init and migration started (users)")
	var err error
	DB, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{Logger: gorm_logrus.New()})
	if err != nil {
		panic("failed to connect database")
	}

	// Миграция схем
	DB.AutoMigrate(&User{})

}
func UpdateLastSeen(id int64) {
	var user User
	DB.First(&user, id)
	user.LastSeen = time.Now()
	DB.Save(&user)
}
func GetUniqueChats() []User {
	var users []User
	DB.Distinct("chatid").Find(&users)
	return users

}
