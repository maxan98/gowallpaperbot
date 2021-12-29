package clients

import "sync"

type clients struct {
	ips  []string
	lock sync.Mutex
}

var instance *clients
var once sync.Once

func GetInstance() *clients {
	once.Do(func() {
		instance = new(clients)
	})
	return instance
}
func (s *clients) GetClients() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ips
}
func (s *clients) AppendClient(ip string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, i := range s.ips {
		if i == ip {
			return
		}
	}
	s.ips = append(s.ips, ip)
}
