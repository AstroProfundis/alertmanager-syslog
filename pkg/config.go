package webhook

import (
	"os"

	"gopkg.in/yaml.v2"
)

type column struct {
	Type      string `yaml:"type"`
	Value     string `yaml:"value,omitempty"`
	Key       string `yaml:"key,omitempty"`
	Numeric   bool   `yaml:"numeric,omitempty"`
	StripPort bool   `yaml:"stripPort,omitempty"`
}

// UnmarshalYAML implement default values of column
func (c *column) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawCol column
	raw := rawCol{
		Numeric:   false,
		StripPort: false,
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}
	*c = column(raw)
	return nil
}

type section struct {
	Join      bool     `yaml:"join"`
	Delimiter string   `yaml:"delimiter,omitempty"`
	Columns   []column `yaml:"columns"`
}

// UnmarshalYAML implement default values of section
func (s *section) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawSec section
	raw := rawSec{
		Join: false,
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}
	*s = section(raw)
	return nil
}

type kvMap struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

type severities struct {
	IncludeResolved bool    `yaml:"includeResolved"`
	ErrorAsEmpty    bool    `yaml:"errorAsEmpty"`
	Type            string  `yaml:"type"`
	Key             string  `yaml:"key"`
	Mode            string  `yaml:"mode"`
	Levels          []kvMap `yaml:"levels"`
}

type custom struct {
	Delimiter         string     `yaml:"delimiter"`
	ReplaceEmpty      string     `yaml:"replaceEmpty,omitempty"`
	ReplaceWhitespace string     `yaml:"replaceWhitespace,omitempty"`
	Severities        severities `yaml:"severities"`
	Sections          []section  `yaml:"sections"`
}

// Config is the output format configurations
type Config struct {
	Mode        string   `yaml:"mode"`
	Severity    string   `yaml:"severity"`
	Facility    string   `yaml:"facility"`
	Labels      []string `yaml:"labels"`
	Annotations []string `yaml:"annotations"`
	Custom      custom   `yaml:"custom"`
}

// LoadConfig loads configs from the specified YAML file
func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		return &Config{}, nil
	}

	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
