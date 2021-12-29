package chats

import "sync"

type chats struct {
	ids  []int64
	lock sync.Mutex
}

var instance *chats
var once sync.Once

func GetInstance() *chats {
	once.Do(func() {
		instance = new(chats)
	})
	return instance
}
func (s *chats) GetClients() []int64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ids
}
func (s *chats) AppendClient(ip int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, i := range s.ids {
		if i == ip {
			return
		}
	}
	s.ids = append(s.ids, ip)
}
