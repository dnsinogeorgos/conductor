package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config stores the configuration loaded during startup.
type Config struct {
	Debug   bool   `json:"debug"`
	Address string `json:"address"`
	Port    int32  `json:"port"`

	PoolName                 string `json:"pool_name" split_words:"true"`
	PoolPath                 string `json:"pool_path" split_words:"true"`
	PoolDev                  string `json:"pool_dev" split_words:"true"`
	FilesystemName           string `json:"filesystem_name" split_words:"true"`
	FilesystemPath           string `json:"filesystem_path" split_words:"true"`
	CastPath                 string `json:"cast_path" split_words:"true"`
	ReplicaPath              string `json:"replica_path" split_words:"true"`
	PortLowerBound           int32  `json:"port_from" split_words:"true"`
	PortUpperBound           int32  `json:"port_to" split_words:"true"`
	MainUnit                 string `json:"main_unit" split_words:"true"`
	ConfigTemplatePath       string `json:"config_template_path" split_words:"true"`
	UnitTemplateString       string `json:"unit_template_string" split_words:"true"`
	ConfigPathTemplateString string `json:"config_path_template_string" split_words:"true"`
}

// NewConfig creates an empty config instance.
func NewConfig(name string) (*Config, error) {
	configfile := flag.String("c", "conductor.json", "path to configuration file")
	flag.Parse()

	config := Config{
		Debug:          false,
		Address:        "127.0.0.1",
		Port:           8080,
		PoolName:       "rootpool",
		PoolPath:       "/rootpool",
		FilesystemName: "rootfs",
		CastPath:       "/rootfs_cast",
		ReplicaPath:    "/rootfs_replica",
	}

	err := envconfig.Process(name, &config)
	if err != nil {
		return &Config{}, err
	}

	err = config.LoadJson(*configfile)
	if err != nil {
		_, ok := err.(*os.PathError)
		if !ok {
			return &Config{}, err
		}
	}

	if config.PoolDev == "" {
		return &Config{}, MissingConfigurationVariableError{t: "string", n: "PoolDev"}
	}

	if config.FilesystemPath == "" {
		return &Config{}, MissingConfigurationVariableError{t: "string", n: "FilesystemPath"}
	}

	if config.PortLowerBound == 0 {
		return &Config{}, MissingConfigurationVariableError{t: "int", n: "PortLowerBound"}
	}

	if config.PortUpperBound == 0 {
		return &Config{}, MissingConfigurationVariableError{t: "int", n: "PortUpperBound"}
	}

	if config.MainUnit == "" {
		return &Config{}, MissingConfigurationVariableError{t: "string", n: "MainUnit"}
	}

	if config.ConfigTemplatePath == "" {
		return &Config{}, MissingConfigurationVariableError{t: "string", n: "ConfigTemplatePath"}
	}

	if config.UnitTemplateString == "" {
		return &Config{}, MissingConfigurationVariableError{t: "string", n: "UnitTemplateString"}
	}

	if config.ConfigPathTemplateString == "" {
		return &Config{}, MissingConfigurationVariableError{t: "string", n: "ConfigTemplatePath"}
	}

	return &config, nil
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
