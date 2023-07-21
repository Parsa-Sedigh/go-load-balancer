package config

import (
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	conf, err := LoadConfig(strings.NewReader(`
strategy: "RoundRobin"
services:
   - name: "test service"
     replicas:
        - "localhost:8081"
        - "localhost:8082"
`))
	if err != nil {
		t.Errorf("Error should be nil '%s'", err)
	}

	if conf.Strategy != "RoundRobin" {
		t.Errorf("Strategy expected to equal 'RoundRobin' got '%s' instead", conf.Strategy)
	}

	if len(conf.Services) != 1 {
		t.Errorf("Expected service count to be 1 got '%d' instead", len(conf.Services))
	}
}
