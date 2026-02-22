package core

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type FrameworkDetector struct {
}

type DetectionResult struct {
	URL          *url.URL
	StatusCode   int
	Server       string
	Frameworks   []string
	Languages    []string
	Headers      map[string]string
	GeneratedBy  string
	XPoweredBy   string
	Cookies      []string
	Technologies []string
}

func NewFrameworkDetector() *FrameworkDetector {
	return &FrameworkDetector{}
}

func (d *FrameworkDetector) Detect(response *http.Response) (*DetectionResult, error) {
	result := &DetectionResult{
		URL:          response.Request.URL,
		StatusCode:   response.StatusCode,
		Server:       response.Header.Get("Server"),
		GeneratedBy:  response.Header.Get("Generated-By"),
		XPoweredBy:   response.Header.Get("X-Powered-By"),
		Headers:      make(map[string]string),
		Frameworks:   make([]string, 0),
		Languages:    make([]string, 0),
		Cookies:      make([]string, 0),
		Technologies: make([]string, 0),
	}

	// copy headers
	for key, values := range response.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
			if key == "Set-Cookie" {
				result.Cookies = append(result.Cookies, values...)
			}
		}
	}

	body := make([]byte, 100000)
	n, _ := response.Body.Read(body)
	bodyStr := string(body[:n])

	d.detectFromBody(result, bodyStr)
	d.detectFromCookies(result)
	d.detectFromHeaders(result)

	d.PrintResults(result)

	return result, nil
}

func (d *FrameworkDetector) detectFromHeaders(result *DetectionResult) {
	server := strings.ToLower(result.Server)

	switch {
	case strings.Contains(server, "apache"):
		result.Technologies = append(result.Technologies, "Apache")
	case strings.Contains(server, "nginx"):
		result.Technologies = append(result.Technologies, "Nginx")
	case strings.Contains(server, "iis"):
		result.Technologies = append(result.Technologies, "IIS")
	case strings.Contains(server, "cloudflare"):
		result.Technologies = append(result.Technologies, "Cloudflare")
	}

	// X-Powered-By detection
	// poweredBy := strings.ToLower(result.XPoweredBy)

}

func (d *FrameworkDetector) detectFromBody(result *DetectionResult, body string) {
	bodyLower := strings.ToLower(body)

	// PHP Frameworks
	if strings.Contains(bodyLower, "wp-content") || strings.Contains(bodyLower, "wp-includes") {
		result.Languages = append(result.Languages, "PHP")
		result.Frameworks = append(result.Frameworks, "wordpress")
	}
	if strings.Contains(bodyLower, "laravel") || strings.Contains(bodyLower, "laravel_session") {
		result.Languages = append(result.Languages, "PHP")
		result.Frameworks = append(result.Frameworks, "laravel")
	}
	if strings.Contains(bodyLower, "ci_session") || strings.Contains(bodyLower, "codeigniter") {
		result.Languages = append(result.Languages, "PHP")
		result.Frameworks = append(result.Frameworks, "CodeIgniter")
	}
	if strings.Contains(bodyLower, "symfony") || strings.Contains(bodyLower, "sf_") {
		result.Languages = append(result.Languages, "PHP")
		result.Frameworks = append(result.Frameworks, "Symfony")
	}

	// Python frameworks
	if strings.Contains(bodyLower, "csrfmiddlewaretoken") || strings.Contains(bodyLower, "django") {
		result.Languages = append(result.Languages, "Python")
		result.Frameworks = append(result.Frameworks, "Django")
	}
	if strings.Contains(bodyLower, "flask") || strings.Contains(bodyLower, "jinja") {
		result.Languages = append(result.Languages, "Python")
		result.Frameworks = append(result.Frameworks, "Flask")
	}
	if strings.Contains(bodyLower, "fastapi") {
		result.Languages = append(result.Languages, "Python")
		result.Frameworks = append(result.Frameworks, "FastAPI")
	}

	// Ruby frameworks
	if strings.Contains(bodyLower, "rails") || strings.Contains(bodyLower, "csrf-param") {
		result.Languages = append(result.Languages, "Ruby")
		result.Frameworks = append(result.Frameworks, "Ruby on Rails")
	}

	// Java frameworks
	if strings.Contains(bodyLower, "javax.faces") || strings.Contains(bodyLower, "jsf") {
		result.Languages = append(result.Languages, "Java")
		result.Frameworks = append(result.Frameworks, "JSF")
	}
	if strings.Contains(bodyLower, "spring") || strings.Contains(bodyLower, "csrf") && strings.Contains(bodyLower, "org.springframework") {
		result.Languages = append(result.Languages, "Java")
		result.Frameworks = append(result.Frameworks, "Spring")
	}

	// Javascript Frameworks
	if strings.Contains(bodyLower, "next/data") || strings.Contains(bodyLower, "_next") {
		result.Languages = append(result.Languages, "JavaScript/TypeScript")
		result.Frameworks = append(result.Frameworks, "Next.js")
		result.Technologies = append(result.Technologies, "React")
	}
	if strings.Contains(bodyLower, "react") || strings.Contains(bodyLower, "reactdom") {
		result.Technologies = append(result.Technologies, "React")
	}
	if strings.Contains(bodyLower, "vue") || strings.Contains(bodyLower, "vuejs") {
		result.Technologies = append(result.Technologies, "Vue.js")
	}
	if strings.Contains(bodyLower, "angular") || strings.Contains(bodyLower, "ng-") {
		result.Technologies = append(result.Technologies, "Angular")
	}
	if strings.Contains(bodyLower, "gatsby") {
		result.Languages = append(result.Languages, "JavaScript/TypeScript")
		result.Frameworks = append(result.Frameworks, "Gatsby")
		result.Technologies = append(result.Technologies, "React")
	}

	// Detect from script tags and meta tags
	scriptRegex := regexp.MustCompile(`<script[^>]*src=["']([^"']*)["']`)
	matches := scriptRegex.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		if len(match) > 1 {
			src := strings.ToLower(match[1])
			if strings.Contains(src, "next") {
				result.Technologies = append(result.Technologies, "Next.js")
			}
			if strings.Contains(src, "react") {
				result.Technologies = append(result.Technologies, "React")
			}
			if strings.Contains(src, "vue") {
				result.Technologies = append(result.Technologies, "Vue.js")
			}
			if strings.Contains(src, "angular") {
				result.Technologies = append(result.Technologies, "Angular")
			}
		}
	}
}

func (d *FrameworkDetector) detectFromCookies(result *DetectionResult) {
	for _, cookie := range result.Cookies {
		cookieLower := strings.ToLower(cookie)

		switch {
		case strings.Contains(cookieLower, "laravel_session"):
			addUnique(&result.Languages, "PHP")
			addUnique(&result.Frameworks, "Laravel")
		case strings.Contains(cookieLower, "ci_session"):
			addUnique(&result.Languages, "PHP")
			addUnique(&result.Frameworks, "CodeIgniter")
		case strings.Contains(cookieLower, "wordpress"):
			addUnique(&result.Languages, "PHP")
			addUnique(&result.Frameworks, "WordPress")
		case strings.Contains(cookieLower, "symfony"):
			addUnique(&result.Languages, "PHP")
			addUnique(&result.Frameworks, "Symfony")
		case strings.Contains(cookieLower, "django"):
			addUnique(&result.Languages, "Python")
			addUnique(&result.Frameworks, "Django")
		case strings.Contains(cookieLower, "flask"):
			addUnique(&result.Languages, "Python")
			addUnique(&result.Frameworks, "Flask")
		case strings.Contains(cookieLower, "rails"):
			addUnique(&result.Languages, "Ruby")
			addUnique(&result.Frameworks, "Ruby on Rails")
		case strings.Contains(cookieLower, "spring"):
			addUnique(&result.Languages, "Java")
			addUnique(&result.Frameworks, "Spring")
		case strings.Contains(cookieLower, "next"):
			addUnique(&result.Languages, "JavaScript/TypeScript")
			addUnique(&result.Frameworks, "Next.js")
			addUnique(&result.Technologies, "React")
		case strings.Contains(cookieLower, "gatsby"):
			addUnique(&result.Languages, "JavaScript/TypeScript")
			addUnique(&result.Frameworks, "Gatsby")
			addUnique(&result.Technologies, "React")
		case strings.Contains(cookieLower, "react"):
			addUnique(&result.Technologies, "React")
		case strings.Contains(cookieLower, "vue"):
			addUnique(&result.Technologies, "Vue.js")
		case strings.Contains(cookieLower, "angular"):
			addUnique(&result.Technologies, "Angular")
		}
	}
}

func (d *FrameworkDetector) PrintResults(result *DetectionResult) {
	fmt.Printf("\n============== Framework Detection Results ==============\n")
	fmt.Printf("URL: %s\n", result.URL)
	fmt.Printf("Status Code: %d\n", result.StatusCode)

	if result.Server != "" {
		fmt.Printf("Server: %s\n", result.Server)
	}

	fmt.Println("\n--- Detected Languages ---")
	if len(result.Languages) == 0 {
		fmt.Println("No languages detected")
	} else {
		for _, lang := range unique(result.Languages) {
			fmt.Printf("  ✓ %s\n", lang)
		}
	}

	fmt.Println("\n--- Detected Frameworks ---")
	if len(result.Frameworks) == 0 {
		fmt.Println("No frameworks detected")
	} else {
		for _, framework := range unique(result.Frameworks) {
			fmt.Printf("  ✓ %s\n", framework)
		}
	}

	fmt.Println("\n--- Detected Technologies ---")
	if len(result.Technologies) == 0 {
		fmt.Println("No technologies detected")
	} else {
		for _, tech := range unique(result.Technologies) {
			fmt.Printf("  ✓ %s\n", tech)
		}
	}

	if result.XPoweredBy != "" {
		fmt.Printf("\nX-Powered-By: %s\n", result.XPoweredBy)
	}

	if result.GeneratedBy != "" {
		fmt.Printf("Generated-By: %s\n", result.GeneratedBy)
	}

	if len(result.Cookies) > 0 {
		fmt.Println("\n--- Cookies ---")
		for _, cookie := range result.Cookies {
			fmt.Printf("  %s\n", cookie)
		}
	}

	fmt.Printf("\n=== Detection Complete ===\n\n")

}

func addUnique(slice *[]string, item string) {
	for _, s := range *slice {
		if s == item {
			return
		}
	}
	*slice = append(*slice, item)
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
