package store

import (
	"encoding/json"
	"os"
	"sync"
)

type Site struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Name   string `json:"name"`
	Method string `json:"method"` // "http", "tcp", "ping"
}

type ProxyConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"` // e.g., "socks5://127.0.0.1:1080"
}

type Data struct {
	Sites []Site      `json:"sites"`
	Proxy ProxyConfig `json:"proxy"`
}

type Store struct {
	mu       sync.RWMutex
	filePath string
	Data     Data
}

func NewStore(filePath string) (*Store, error) {
	s := &Store{
		filePath: filePath,
		Data: Data{
			Sites: []Site{},
			Proxy: ProxyConfig{
				Enabled: false,
				URL:     "",
			},
		},
	}
	if err := s.Load(); err != nil {
		if os.IsNotExist(err) {
			// Create default if not exists
			return s, s.Save()
		}
		return nil, err
	}
	return s, nil
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, &s.Data)
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.Data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Store) AddSite(site Site) error {
	s.mu.Lock()
	s.Data.Sites = append(s.Data.Sites, site)
	s.mu.Unlock()
	return s.Save()
}

func (s *Store) UpdateSite(updated Site) error {
	s.mu.Lock()
	for i, site := range s.Data.Sites {
		if site.ID == updated.ID {
			s.Data.Sites[i] = updated
			break
		}
	}
	s.mu.Unlock()
	return s.Save()
}

func (s *Store) RemoveSite(id string) error {
	s.mu.Lock()
	newSites := []Site{}
	for _, site := range s.Data.Sites {
		if site.ID != id {
			newSites = append(newSites, site)
		}
	}
	s.Data.Sites = newSites
	s.mu.Unlock()
	return s.Save()
}

func (s *Store) UpdateProxy(config ProxyConfig) error {
	s.mu.Lock()
	s.Data.Proxy = config
	s.mu.Unlock()
	return s.Save()
}

func (s *Store) GetSites() []Site {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Data.Sites
}

func (s *Store) GetProxy() ProxyConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Data.Proxy
}
