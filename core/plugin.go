package core

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

type PluginConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Severity    string `yaml:"severity"`

	Framework  string `yaml:"framework"`
	Technology string `yaml:"technology"`
	Language   string `yaml:"language"`

	Headers map[string]string
	Method  string
	Timeout int

	Payloads []PayloadConfig `yaml:"payloads"`
	Matchers struct {
		Name         string `yaml:"name"`
		Type         string `yaml:"type"`
		ExtractRegex string `yaml:"extract_regex"`
		DecodeURL    bool   `yaml:"decode_url"`
		DecodePipes  bool   `yaml:"decode_pipes"`
	} `yaml:"matchers"`
}

type PayloadConfig struct {
	Name               string `yaml:"name"`
	CommandPlaceholder string `yaml:"command_placeholder"`
	JsonTemplate       string `yaml:"json_template"`
	MultipartFormData  []struct {
		Name    string `yaml:"name"`
		Content string `yaml:"content"`
	} `yaml:"multipart_form_data"`
}

type ExecutionContext struct {
	Command        string
	EscapedCommand string
	Boundary       string
	RequestID      string
	JsonTemplate   string
	CommandResult  string
}

type PluginEngine struct {
	config  *PluginConfig
	client  *http.Client
	context *ExecutionContext
}

func NewPluginEngine(configPath string, cfg *Config) (*PluginEngine, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read plugin config: %v", err)
	}
	var config PluginConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Failed to parse YAML %v", err)
	}

	var client *http.Client

	if cfg.Proxy.Enabled {
		proxyAddress := fmt.Sprintf("http://%s:%d", cfg.Proxy.Address, cfg.Proxy.Port)
		fmt.Printf("Using proxy: %s....\n", proxyAddress)
		proxyUrl, _ := url.Parse(proxyAddress)
		tr := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		client = &http.Client{Timeout: time.Duration(config.Timeout) * time.Second, Transport: tr}
	} else {
		client = &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	}

	return &PluginEngine{
		config:  &config,
		client:  client,
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

func (e *PluginEngine) Execute(target DetectionResult) error {
	fmt.Printf("Executing plugin: %s on host %s\n", e.config.Name, target.URL.String())

	// Check if framework, languages and Technologies meet the url
	if !e.isTargetCompatible(target) {
		fmt.Printf("Target %s is not compatible with plugin \n", target.URL.String())
		return fmt.Errorf("Target %s is not compatible with plugin", target.URL.String())
	}

	// Use first payload - Will be extended to support multiple
	payload := e.config.Payloads[0]

	// Build multipart payload
	multipartData, err := e.BuildMultiPartPayload(&payload)

	if err != nil {
		return err
	}

	// send request
	resp, err := e.SendRequest(target.URL.String(), multipartData)
	if err != nil {
		fmt.Printf("An error occured: %v\n", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status Code: %s\n", resp.Status)

	result, err := e.ExtractResult(resp)
	if err != nil {
		fmt.Printf("Error getting result: %v\n", err)
		return err
	}

	fmt.Printf("[+] Result %s\n", result)

	return nil
}

func (e *PluginEngine) ExecuteTemplate(tmplStr string) (string, error) {
	tmpl, err := template.New("template").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, e.context); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (e *PluginEngine) BuildMultiPartPayload(payload *PayloadConfig) ([]byte, error) {
	var multipartBody bytes.Buffer

	jsonTemplate, err := e.ExecuteTemplate(payload.JsonTemplate)

	if err != nil {
		return nil, err
	}

	e.context.JsonTemplate = jsonTemplate

	for _, part := range payload.MultipartFormData {
		// Execute part content template
		content, err := e.ExecuteTemplate(part.Content)
		if err != nil {
			fmt.Printf("Error: %v\n", err)

			return nil, err
		}

		// Write multipart boundary and headers with CRLF
		multipartBody.WriteString(fmt.Sprintf("------WebKitFormBoundaryx8jO2oVc6SWP3Sad\r\n"))
		multipartBody.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"\r\n", part.Name))
		multipartBody.WriteString("\r\n")
		multipartBody.WriteString(content)
		multipartBody.WriteString("\r\n")
	}

	// Write final boundary
	multipartBody.WriteString(fmt.Sprintf("------WebKitFormBoundaryx8jO2oVc6SWP3Sad--\r\n"))

	return multipartBody.Bytes(), nil

}

func (e *PluginEngine) ExtractResult(resp *http.Response) (string, error) {
	headerValue := resp.Header.Get(e.config.Matchers.Name)
	if headerValue == "" {
		return "", fmt.Errorf("header %s not found", e.config.Matchers.Name)
	}

	// Extract with regex
	re := regexp.MustCompile(e.config.Matchers.ExtractRegex)
	matches := re.FindStringSubmatch(headerValue)
	if len(matches) > 2 {
		return "", fmt.Errorf("failed to extract result from header")
	}

	result := matches[1]

	// Url decode if configured
	if e.config.Matchers.DecodeURL {
		decoded, err := url.QueryUnescape(result)
		if err == nil {
			result = decoded
		}
	}

	// Decode pipes if configured
	if e.config.Matchers.DecodePipes {
		result = strings.ReplaceAll(result, " | ", "\n")
	}

	fmt.Printf("Result %s \n", result)
	return result, nil
}

func (e *PluginEngine) isTargetCompatible(target DetectionResult) bool {
	// if no requirements specified, assumes it does not meet them.
	if e.config.Framework == "" && e.config.Technology == "" && e.config.Language == "" {
		return false
	}

	// Check frameworks
	if !e.stringInSlice(e.config.Framework, target.Frameworks) {
		return false
	}

	if !e.stringInSlice(e.config.Technology, target.Technologies) {
		return false
	}

	if !e.stringInSlice(e.config.Language, target.Languages) {
		return false
	}

	return true
}

func (e *PluginEngine) stringInSlice(search string, slice []string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, search) {
			return true
		}
	}
	return false
}
