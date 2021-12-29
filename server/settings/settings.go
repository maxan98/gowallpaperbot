package settings

import (
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type settings struct {
	Token      string  `yaml:"token"`
	Filepath   string  `yaml:"fileDir"`
	AllowedIDs []int64 `yaml:"allowedIDs"`
	LogFile    string  `yaml:"logFile"`
	lock       sync.RWMutex
}

var instance *settings
var once sync.Once

func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func (s *settings) ValidateConfig() bool {
	if s.LogFile == "" || s.AllowedIDs == nil || s.Filepath == "" || s.Token == "" {
		return false
	}
	return true
}

func ParseFlags() (string, error) {
	var configPath string
	flag.StringVar(&configPath, "c", "./config.yaml", "path to config file")
	flag.Parse()

	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}

func NewConfig(configPath string) (*settings, error) {
	set := GetInstance()
	set.lock.Lock()
	defer set.lock.Unlock()
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)
	if err := d.Decode(&set); err != nil {
		return nil, err
	}
	ok := set.ValidateConfig()
	if !ok {
		return nil, errors.New("invalid Config")
	}
	return set, nil
}

func GetInstance() *settings {
	once.Do(func() {
		instance = new(settings)
	})
	return instance
}
func (s *settings) GetSettings() *settings {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s
}
func (s *settings) GetToken() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.Token
}

func (s *settings) GetFilePath() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.Filepath
}

func (s *settings) SetToken(token string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Token = token
}
func (s *settings) SetFilePath(filePath string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Filepath = filePath
}
