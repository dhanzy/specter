package core

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type PluginConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Severity    string `yaml:"severity"`

	Headers map[string]string
	Method  string
	Timeout int
}

type ExecutionContext struct {
	Command        string
	EscapedCommand string
	RequestID      string
}

type PluginEngine struct {
	config  *PluginConfig
	client  *http.Client
	context *ExecutionContext
}

func NewPluginEngine(configPath string) (*PluginEngine, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read plugin config: %v", err)
	}
	var config PluginConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Failed to parse YAML %v", err)
	}

	return &PluginEngine{
		config:  &config,
		client:  &http.Client{Timeout: time.Duration(config.Timeout) * time.Second},
		context: &ExecutionContext{},
	}, nil
}

func (e *PluginEngine) SendRequest(targetURL string, payload []byte) (*http.Response, error) {

	headers := make(map[string]string)

	req, err := http.NewRequest(e.config.Method, targetURL, bytes.NewReader(payload))

	for key, value := range e.config.Headers {
		req.Header.Set(key, value)
		headers[key] = value
	}

	if err != nil {
		return nil, err
	}

	return e.client.Do(req)
}

func (e *PluginEngine) Execute(targetURL string) {
	fmt.Printf("Executing plugin: %s\n", e.config.Name)
}
