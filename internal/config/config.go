package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config stores the configuration loaded during startup.
type Config struct {
	ServiceName    string `json:"service_name"`
	PoolName       string `json:"pool_name"`
	PoolPath       string `json:"pool_path"`
	PoolDev        string `json:"pool_dev"`
	FsName         string `json:"fs_name"`
	FsPath         string `json:"fs_path"`
	CastPath       string `json:"cast_path"`
	ReplicaPath    string `json:"replica_path"`
	PortLowerBound uint16 `json:"port_from"`
	PortUpperBound uint16 `json:"port_to"`
}

// NewConfig creates an empty config instance.
func NewConfig(p string) (*Config, error) {
	c := Config{}
	err := c.LoadJson(p)
	if err != nil {
		return &Config{}, err
	}

	return &c, nil
}

// LoadJson loads the json values to the config instance.
func (c *Config) LoadJson(p string) error {
	file, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		return err
	}

	return nil
}
