package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	Labels      []string `yaml:"labels"`
	Annotations []string `yaml:"annotations"`
}

// loadConfig loads configs from the specified YAML file
func loadConfig(filename string) (*config, error) {
	if filename == "" {
		return &config{}, nil
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg config
	err = yaml.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
