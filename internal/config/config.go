package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Db_url            string `json:db_url`
	Current_user_name string `json:current_user_name`
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c *Config) SetUser(name string) error {

	filepath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	conf := Config{c.Db_url, name}
	jsonConf, err := json.Marshal(&conf)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, jsonConf, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	const configFileName = ".gatorconfig.json"
	homepath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(homepath)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == configFileName {
			return filepath.Join(homepath, configFileName), nil
		}
	}
	return "", errors.New("no config file found in ~")
}
