package auth

import (
	"encoding/json"
	"os"
	"path/filepath"

	"jeeves/config"
)

type Session struct {
	Token string `json:"token"`
}

func sessionPath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

func Save(token string) error {
	if err := config.EnsureDir(); err != nil {
		return err
	}
	path, err := sessionPath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(Session{Token: token})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func Load() (*Session, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func Clear() error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func Token() string {
	s, err := Load()
	if err != nil || s == nil {
		return ""
	}
	return s.Token
}
