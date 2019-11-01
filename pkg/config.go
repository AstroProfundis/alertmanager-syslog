package webhook

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type section struct {
	Type    string `yaml:"type"`
	Value   string `yaml:"value,omitempty"`
	Key     string `yaml:"key,omitempty"`
	Numeric bool   `yaml:"numeric,omitempty"`
}

type kvMap struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

type severities struct {
	IncludeResolved bool    `yaml:"includeResolved"`
	Key             string  `yaml:"key"`
	Mode            string  `yaml:"mode"`
	Levels          []kvMap `yaml:"levels"`
}

type custom struct {
	Enabled    bool       `yaml:"enabled"`
	Delimiter  string     `yaml:"delimiter"`
	Severities severities `yaml:"severities"`
	Sections   []section  `yaml:"sections"`
}

type config struct {
	Mode        string   `yaml:"mode"`
	Severity    string   `yaml:"severity"`
	Facility    string   `yaml:"facility"`
	Labels      []string `yaml:"labels"`
	Annotations []string `yaml:"annotations"`
	Custom      custom   `yaml:"custom"`
}

// LoadConfig loads configs from the specified YAML file
func LoadConfig(filename string) (*config, error) {
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
