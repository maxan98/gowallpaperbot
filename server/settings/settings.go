package settings

import "sync"

type settings struct {
	token    string
	filepath string
	lock     sync.RWMutex
}

var instance *settings
var once sync.Once

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
	return s.token
}

func (s *settings) GetFilePath() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.filepath
}

func (s *settings) SetToken(token string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.token = token
}
func (s *settings) SetFilePath(filePath string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.filepath = filePath
}
