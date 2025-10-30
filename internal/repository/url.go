package repository

type URLStore struct {
	urls map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{
		urls: make(map[string]string),
	}
}

func (s *URLStore) SaveWithID(id, url string) {
	s.urls[id] = url
}

func (s *URLStore) Get(id string) (string, bool) {
	url, ok := s.urls[id]
	return url, ok
}
