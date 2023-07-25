package config

import (
	"gopkg.in/yaml.v2"
	"io"
)

func LoadConfig(reader io.Reader) (*Config, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var conf Config

	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
