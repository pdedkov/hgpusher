package main

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

// Config app config
type Config struct {
	Login    string
	Password string
	Username string
	Root     string
}

// NewConfigFromFile load config from file
func NewConfigFromFile(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("Empty path")
	}

	// check config file
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("File '%s' does not exists", path)
		}
		return nil, err
	}

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return nil, fmt.Errorf("Error while loading config file: %v", err)
	}

	return &conf, nil
}
